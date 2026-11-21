// Pinned direct helper with synthetic invocation and a credential-refusing host.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Newtonsoft.Json;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;

namespace TeamsContractOracle
{
    public static class ClientConfigurationOracle
    {
        public static ProviderCapture Run(string operation, string input, int status, string response, int forceMode)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null); var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract; var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status, response); using var client = new HttpClient(sender);
            client.DefaultRequestHeaders.Add("MPACmdlet", "true");
            clientField.SetValue(null, client); versionField.SetValue(null, "8.1.0"); HostSingletionInstance.HostContract = host;
            try {
                const string identity = "scope /Fixture / Child", token = "offline-fixture-only", origin = "api.interfaces.records.teams.microsoft.com";
                string command = operation == "patch" ? "Set-CsTeamsClientConfiguration" : "Get-CsTeamsClientConfiguration";
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (forceMode >= 0) bound.Add("Force", new SwitchParameter(forceMode == 1));
                var value = JsonConvert.DeserializeObject<TeamsClientConfiguration>(input);
                value.Identity = "read-model-identity";
                var requested = JsonConvert.DeserializeObject<Dictionary<string, object>>(input);
                if (operation == "patch") {
                    foreach (var name in requested.Keys) {
                        if (name == "Identity") throw new ArgumentException("Identity is selected separately");
                        var property = typeof(TeamsClientConfiguration).GetProperty(name);
                        if (property == null) throw new ArgumentException("Unknown synthetic setting");
                        bound.Add(name, property.GetValue(value));
                    }
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetTeamsClientConfigurationCmdlet)), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsClientConfiguration>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsClientConfiguration>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsClientConfiguration>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    _ => throw new ArgumentException("Unknown synthetic client configuration operation")
                };
                sender.Capture.Success = result.IsSuccessStatusCode; sender.Capture.DecodedNull = result.PolicyResult == null;
                sender.Capture.DecodedType = result.PolicyResult?.GetType().FullName; sender.Capture.Warnings = result.Warnings?.ToArray();
            }
            catch (Exception error) { sender.Capture.FailureType = error.GetType().FullName; }
            finally { sender.Capture.TokenRequests = host.TokenRequests; clientField.SetValue(null, oldClient); versionField.SetValue(null, oldVersion); HostSingletionInstance.HostContract = oldHost; }
            return sender.Capture;
        }
        public static string Model(string input) {
            object value = null; string failure = null;
            try {value = input == "constructor" ? new TeamsClientConfiguration() : JsonConvert.DeserializeObject<TeamsClientConfiguration>(input);}
            catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new {failure,value});
        }
    }
}
