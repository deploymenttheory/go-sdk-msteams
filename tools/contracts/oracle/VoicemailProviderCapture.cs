// Offline compiled CRUD helpers. Generated assignment is captured separately.
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
    public static class VoicemailOracle
    {
        public static ProviderCapture Run(string operation, int status, string response, bool force, string duration)
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
                string command = operation switch {
                    "create" => "New-CsOnlineVoicemailPolicy", "patch" => "Set-CsOnlineVoicemailPolicy", "delete" => "Remove-CsOnlineVoicemailPolicy",
                    _ => "Get-CsOnlineVoicemailPolicy"
                };
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (force) bound.Add("Force", new SwitchParameter(false));
                var policy = new OnlineVoicemailPolicy {
                    Identity = "read-model-identity", Description = "", ShareData = "FutureValue", PrimarySystemPromptLanguage = "FuturePrimary", SecondarySystemPromptLanguage = "FutureSecondary",
                    PreambleAudioFile = "fixture-audio-reference", PostambleAudioFile = "", EnableTranscription = false, EnableTranscriptionProfanityMasking = false,
                    EnableEditingCallAnswerRulesSetting = false, EnableTranscriptionTranslation = false, PreamblePostambleMandatory = false, EnableVoicemailTriage = false,
                    MaximumRecordingLength = duration == "null" ? (TimeSpan?)null : TimeSpan.FromTicks(duration == "tick" ? 1 : 900000000)
                };
                if (operation == "create" || operation == "patch") {
                    foreach (var property in typeof(OnlineVoicemailPolicy).GetProperties().Where(p => p.Name != "Identity"))
                        bound.Add(property.Name, property.GetValue(policy));
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetOnlineVoicemailPolicyCmdlet)), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<OnlineVoicemailPolicy>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<OnlineVoicemailPolicy>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<OnlineVoicemailPolicy>.CreateAsync(policy, identity, token, origin, invocation, null).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<OnlineVoicemailPolicy>.SetAsync(policy, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<OnlineVoicemailPolicy>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    _ => throw new ArgumentException("Unknown synthetic operation")
                };
                sender.Capture.Success = result.IsSuccessStatusCode; sender.Capture.DecodedNull = result.PolicyResult == null;
                sender.Capture.DecodedType = result.PolicyResult?.GetType().FullName; sender.Capture.Warnings = result.Warnings?.ToArray();
            }
            catch (Exception error) { sender.Capture.FailureType = error.GetType().FullName; }
            finally {
                sender.Capture.TokenRequests = host.TokenRequests; HostSingletionInstance.HostContract = oldHost;
                versionField.SetValue(null, oldVersion); clientField.SetValue(null, oldClient);
            }
            return sender.Capture;
        }
    }
}
