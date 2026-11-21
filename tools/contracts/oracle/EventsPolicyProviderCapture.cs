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
    public static class EventsPolicyOracle
    {
        public static ProviderCapture Run(string operation, string input, int status, string response, int forceMode, bool nullPolicyName, int rank)
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
                string command = operation switch { "create" => "New-CsTeamsEventsPolicy", "patch" => "Set-CsTeamsEventsPolicy", "delete" => "Remove-CsTeamsEventsPolicy", "user" or "tenant" or "group" => "Grant-CsTeamsEventsPolicy", _ => "Get-CsTeamsEventsPolicy" };
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (forceMode >= 0) bound.Add("Force", new SwitchParameter(forceMode == 1));
                var value = JsonConvert.DeserializeObject<TeamsEventsPolicy>(input);
                value.Identity = "read-model-identity";
                var requested = JsonConvert.DeserializeObject<Dictionary<string, object>>(input);
                if (operation == "create" || operation == "patch") {
                    foreach (var name in requested.Keys) {
                        if (name == "Identity") throw new ArgumentException("Identity is selected separately");
                        var property = typeof(TeamsEventsPolicy).GetProperty(name);
                        if (property == null) throw new ArgumentException("Unknown synthetic setting");
                        bound.Add(name, property.GetValue(value));
                    }
                }
                string policyName = nullPolicyName ? null : "";
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, operation switch {"create" => typeof(NewTeamsEventsPolicyCmdlet), "patch" => typeof(SetTeamsEventsPolicyCmdlet), "delete" => typeof(RemoveTeamsEventsPolicyCmdlet), "user" or "tenant" or "group" => typeof(GrantTeamsEventsPolicyCmdlet), _ => typeof(GetTeamsEventsPolicyCmdlet)}), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsEventsPolicy>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsEventsPolicy>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<TeamsEventsPolicy>.CreateAsync(value, identity, token, origin, invocation, null).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<TeamsEventsPolicy>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsEventsPolicy>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "user" => HttpRequestHelper<TeamsEventsPolicy>.GrantToUserAsync("user/fixture", policyName, token, origin, invocation).GetAwaiter().GetResult(),
                    "tenant" => HttpRequestHelper<TeamsEventsPolicy>.GrantToTenantAsync(policyName, token, origin, invocation, forceMode == 1).GetAwaiter().GetResult(),
                    "group" => HttpRequestHelper<TeamsEventsPolicy>.GrantToGroupAsync("group/fixture", policyName, token, origin, invocation, rank).GetAwaiter().GetResult(),
                    _ => throw new ArgumentException("Unknown synthetic events policy operation")
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
            try {value = input == "constructor" ? new TeamsEventsPolicy() : JsonConvert.DeserializeObject<TeamsEventsPolicy>(input);}
            catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new {failure,value});
        }
    }
}
