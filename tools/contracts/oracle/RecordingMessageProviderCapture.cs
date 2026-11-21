// Development-only pinned helper and scope-preparation observations. The memory
// sender is installed before any command constructor; no lifecycle/auth runs.
using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;
using Newtonsoft.Json;

namespace TeamsContractOracle
{
    public sealed class RecordingRequestCapture
    {
        public ProviderCapture Request { get; set; }
        public string PreparedIdentity { get; set; }
        public string RelativeId { get; set; }
        public bool? ConstructorIdWasGuid { get; set; }
        public string Metadata { get; set; }
    }

    public static class RecordingMessageOracle
    {
        public static string Model(string name)
        {
            object result = null; string failure = null;
            try {
                result = name switch {
                    "language-null" => new RecordingAndTranscriptionLocalizationCustomMessage(null),
                    "language-empty" => new RecordingAndTranscriptionLocalizationCustomMessage(""),
                    "language-space" => new RecordingAndTranscriptionLocalizationCustomMessage(" "),
                    "language-future" => new RecordingAndTranscriptionLocalizationCustomMessage("Future-Language"),
                    "localization-absent" => JsonConvert.DeserializeObject<RecordingAndTranscriptionLocalizationCustomMessage>("{}"),
                    "localization-null-language" => JsonConvert.DeserializeObject<RecordingAndTranscriptionLocalizationCustomMessage>("{\"Language\":null}"),
                    "message-absent" => new TeamsRecordingAndTranscriptionCustomMessage(),
                    "message-null-id" => new TeamsRecordingAndTranscriptionCustomMessage { Id = null },
                    "message-empty-id" => new TeamsRecordingAndTranscriptionCustomMessage { Id = "" },
                    "message-invalid-id" => new TeamsRecordingAndTranscriptionCustomMessage { Id = "fixture-invalid" },
                    "message-id-final-lf" => new TeamsRecordingAndTranscriptionCustomMessage { Id = "00000000-0000-4000-8000-000000000031\n" },
                    "message-null-collection" => JsonConvert.DeserializeObject<TeamsRecordingAndTranscriptionCustomMessage>("{\"RecordingAndTranscriptionLocalizationCustomMessage\":null}"),
                    "message-long-description" => new TeamsRecordingAndTranscriptionCustomMessage { Description = new string('x', 2000) },
                    _ => throw new ArgumentException("Unknown synthetic model")
                };
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, value = result });
        }

        static RecordingAndTranscriptionLocalizationCustomMessage Element(string language, string text, string identity)
        {
            return new RecordingAndTranscriptionLocalizationCustomMessage(language) { InitiatorImplicit = text, Identity = identity };
        }

        public static string List(string name)
        {
            var a = Element("en-US", "a", "fixture-first");
            var duplicate = Element("en-US", "a", "fixture-second");
            var lower = Element("en-us", "a", "fixture-lower");
            var differentText = Element("en-US", "b", "fixture-different");
            var values = new List<RecordingAndTranscriptionLocalizationCustomMessage> { a, duplicate, lower, differentText };
            string failure = null;
            try {
                PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage> edit = name switch {
                    "null" => null,
                    "empty-operators" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new Hashtable { { "Add", Array.Empty<RecordingAndTranscriptionLocalizationCustomMessage>() }, { "Remove", Array.Empty<RecordingAndTranscriptionLocalizationCustomMessage>() } }),
                    "replace-empty" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(Array.Empty<object>()),
                    "replace-duplicates" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new object[] { duplicate, duplicate, lower }),
                    "replace-null" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new object[] { null }),
                    "remove-one" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new Hashtable { { "Remove", new[] { duplicate } } }),
                    "remove-two" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new Hashtable { { "Remove", new[] { duplicate, duplicate } } }),
                    "add-existing" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new Hashtable { { "Add", new[] { duplicate, duplicate } } }),
                    "remove-and-add" => new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(new Hashtable { { "Remove", new[] { duplicate } }, { "Add", new[] { duplicate } } }),
                    _ => throw new ArgumentException("Unknown synthetic list")
                };
                edit.ApplyTo(values, (RecordingAndTranscriptionLocalizationCustomMessage left, RecordingAndTranscriptionLocalizationCustomMessage right) =>
                    left.Language == right.Language && left.InitiatorImplicit == right.InitiatorImplicit && left.ParticipantImplicit == right.ParticipantImplicit &&
                    left.InitiatorExplicit == right.InitiatorExplicit && left.ParticipantExplicitRequested == right.ParticipantExplicitRequested &&
                    left.ParticipantExplicitProvided == right.ParticipantExplicitProvided && left.AgreementDialogue == right.AgreementDialogue);
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, values });
        }

        public static RecordingRequestCapture Run(string operation, string selector, string fields, int status, string response, bool force)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null); var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract; var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status, response); using var client = new HttpClient(sender);
            client.DefaultRequestHeaders.Add("MPACmdlet", "true");
            clientField.SetValue(null, client); versionField.SetValue(null, "8.1.0"); HostSingletionInstance.HostContract = host;
            var capture = new RecordingRequestCapture { Request = sender.Capture };
            try {
                const string relative = "00000000-0000-4000-8000-000000000031", token = "offline-fixture-only", origin = "api.interfaces.records.teams.microsoft.com";
                string id = selector == "invalid-id" ? "fixture-non-guid" : relative;
                string identity = selector == "identity" ? "scope /" + relative + "/ Child" : null;
                object cmdlet = operation switch {
                    "create" => new NewTeamsRecordingAndTranscriptionCustomMessageCmdlet(),
                    "patch" => new SetTeamsRecordingAndTranscriptionCustomMessageCmdlet(),
                    "delete" => new RemoveTeamsRecordingAndTranscriptionCustomMessageCmdlet(),
                    _ => new GetTeamsRecordingAndTranscriptionCustomMessageCmdlet()
                };
                var type = cmdlet.GetType();
                if (operation == "create") {
                    capture.ConstructorIdWasGuid = Guid.TryParse((string)type.GetProperty("Id").GetValue(cmdlet), out _);
                    // Fix the constructor-generated ID for deterministic auto-ID
                    // observations; the parameter remains unbound in that case.
                    type.GetProperty("Id").SetValue(cmdlet, id);
                }
                if (operation != "list") {
                    if (selector == "identity") type.GetProperty("Identity").SetValue(cmdlet, identity);
                    else type.GetProperty("Id").SetValue(cmdlet, id);
                }
                type.GetMethod("PrepareForScopeCheck", BindingFlags.NonPublic | BindingFlags.Instance).Invoke(cmdlet, null);
                identity = (string)type.GetProperty("Identity").GetValue(cmdlet);
                capture.PreparedIdentity = identity;
                capture.RelativeId = (string)type.GetProperty("Id").GetValue(cmdlet);
                var metadata = (Dictionary<string, Dictionary<string, object>>)type.GetProperty("CmdletMetadata").GetValue(cmdlet);
                capture.Metadata = JsonConvert.SerializeObject(metadata);
                if (!Object.ReferenceEquals(clientField.GetValue(null), client)) throw new InvalidOperationException("Memory client replaced");

                string command = operation switch {
                    "create" => "New-CsTeamsRecordingAndTranscriptionCustomMessage",
                    "patch" => "Set-CsTeamsRecordingAndTranscriptionCustomMessage",
                    "delete" => "Remove-CsTeamsRecordingAndTranscriptionCustomMessage",
                    _ => "Get-CsTeamsRecordingAndTranscriptionCustomMessage"
                };
                var bound = new Dictionary<string, object>();
                if (operation != "list" && selector == "identity") bound.Add("Identity", identity);
                else if (operation != "list" && selector != "auto") bound.Add("Id", id);
                if (force) bound.Add("Force", new SwitchParameter(false));
                var policy = new TeamsRecordingAndTranscriptionCustomMessage { Identity = "read-model-identity", Description = "" };
                if (operation != "create") policy.Id = relative;
                policy.RecordingAndTranscriptionLocalizationCustomMessage.Add(new RecordingAndTranscriptionLocalizationCustomMessage("en-US") {
                    InitiatorImplicit = "fixture initiator", ParticipantImplicit = "", InitiatorExplicit = null,
                    ParticipantExplicitRequested = "fixture requested", ParticipantExplicitProvided = null, AgreementDialogue = "fixture agreement"
                });
                if ((operation == "create" || operation == "patch") && fields != "none") {
                    bound.Add("Description", policy.Description);
                    if (fields == "all") bound.Add("RecordingAndTranscriptionLocalizationCustomMessage", new PSListModifier<RecordingAndTranscriptionLocalizationCustomMessage>(policy.RecordingAndTranscriptionLocalizationCustomMessage.ToArray()));
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, type), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsRecordingAndTranscriptionCustomMessage>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsRecordingAndTranscriptionCustomMessage>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<TeamsRecordingAndTranscriptionCustomMessage>.CreateAsync(policy, identity, token, origin, invocation, metadata).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsRecordingAndTranscriptionCustomMessage>.SetAsync(policy, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<TeamsRecordingAndTranscriptionCustomMessage>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
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
            return capture;
        }
    }
}
