// Pinned direct helper with synthetic invocation and a credential-refusing host.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;

namespace TeamsContractOracle
{
    public static class PrivacyOracle
    {
        public static ProviderCapture Run(string operation, int status, string response, bool force, bool nullLocation)
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
                string command = operation == "patch" ? "Set-CsPrivacyConfiguration" : "Get-CsPrivacyConfiguration";
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (force) bound.Add("Force", new SwitchParameter(false));
                var value = new PrivacyConfiguration { Identity = "read-model-identity", EnablePrivacyMode = false, AutoInitiateContacts = true,
                    PublishLocationDataDefault = nullLocation ? null : false, DisplayPublishedPhotoDefault = false };
                if (operation == "patch") {
                    bound.Add("EnablePrivacyMode", false); bound.Add("AutoInitiateContacts", true);
                    bound.Add("PublishLocationDataDefault", value.PublishLocationDataDefault); bound.Add("DisplayPublishedPhotoDefault", false);
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetPrivacyConfigurationCmdlet)), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<PrivacyConfiguration>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<PrivacyConfiguration>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<PrivacyConfiguration>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    _ => throw new ArgumentException("Unknown synthetic privacy operation")
                };
                sender.Capture.Success = result.IsSuccessStatusCode; sender.Capture.DecodedNull = result.PolicyResult == null;
                sender.Capture.DecodedType = result.PolicyResult?.GetType().FullName; sender.Capture.Warnings = result.Warnings?.ToArray();
            }
            catch (Exception error) { sender.Capture.FailureType = error.GetType().FullName; }
            finally { sender.Capture.TokenRequests = host.TokenRequests; clientField.SetValue(null, oldClient); versionField.SetValue(null, oldVersion); HostSingletionInstance.HostContract = oldHost; }
            return sender.Capture;
        }
    }
}
