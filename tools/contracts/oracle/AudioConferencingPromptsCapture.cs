// Development-only observations of pinned provider, models and list helpers.
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
    public static class AudioConferencingPromptsOracle
    {
        const string First = "11111111-1111-1111-1111-111111111111";
        const string Second = "22222222-2222-2222-2222-222222222222";

        public static ProviderCapture Run(string operation, string fields, int status, string response, int forceMode)
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
                string command = operation == "patch" ? "Set-CsTeamsAudioConferencingCustomPromptsConfiguration" : "Get-CsTeamsAudioConferencingCustomPromptsConfiguration";
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (forceMode >= 0) bound.Add("Force", new SwitchParameter(forceMode == 1));
                var value = new TeamsAudioConferencingCustomPromptsConfiguration { Identity = "read-model-identity" };
                if (fields != "empty-lists") {
                    value.Prompts.Add(new CustomPrompt(First, "FutureType", "xx-future") { Identity = "prompt-meta", Name = "", AudioPrompt = null, TextPrompt = "text" });
                    var package = new CustomPromptPackage(Second, "") { Identity = "package-meta", Name = null };
                    package.InboundStartRecordingPrompt.AddRange(new[] { First, First, null });
                    package.InboundEndRecordingPrompt.Add("");
                    package.OutboundStartRecordingPrompt.Add("future-reference");
                    package.OutboundEndRecordingPrompt.Add(First);
                    package.MeetingStartRecordingPrompt.Add(Second);
                    value.Packages.Add(package);
                }
                if (operation == "patch") {
                    if (fields != "packages") bound.Add("Prompts", new PSListModifier<CustomPrompt>(value.Prompts.ToArray()));
                    if (fields != "prompts") bound.Add("Packages", new PSListModifier<CustomPromptPackage>(value.Packages.ToArray()));
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetTeamsAudioConferencingCustomPromptsConfigurationCmdlet)), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsAudioConferencingCustomPromptsConfiguration>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsAudioConferencingCustomPromptsConfiguration>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsAudioConferencingCustomPromptsConfiguration>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult(),
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

        public static string Model(string kind, string input)
        {
            object value = null; string failure = null;
            try {
                Type type = kind switch {
                    "configuration" => typeof(TeamsAudioConferencingCustomPromptsConfiguration),
                    "prompt" => typeof(CustomPrompt), "package" => typeof(CustomPromptPackage),
                    _ => throw new ArgumentException("Unknown model")
                };
                value = input == "constructor" ? Activator.CreateInstance(type) : JsonConvert.DeserializeObject(input, type);
            } catch (Exception error) { failure = (error.InnerException ?? error).GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, value });
        }

        public static string Constructor(string kind, string id, string typeOrLocale, string locale, bool nullID, bool nullTypeOrLocale, bool nullLocale)
        {
            object value = null; string failure = null;
            try {
                // Select true null inside C#; the PowerShell string binder uses empty.
                if (nullID) id = null; if (nullTypeOrLocale) typeOrLocale = null; if (nullLocale) locale = null;
                value = kind == "prompt" ? (object)new CustomPrompt(id, typeOrLocale, locale) : new CustomPromptPackage(id, typeOrLocale);
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, value });
        }

        static CustomPrompt Prompt(string name = "a") => new CustomPrompt(First, "FutureType", "xx-future") { Identity = "ignored", Name = name, AudioPrompt = "audio", TextPrompt = "text" };
        static CustomPromptPackage Package(string name = "a") => new CustomPromptPackage(First, "xx-future") { Identity = "ignored", Name = name };

        public static string Lists(string kind, string name)
        {
            if (kind == "prompt") {
                var a = Prompt(); var other = Prompt("b"); var removed = Prompt();
                if (name.StartsWith("different-")) {
                    string field = name.Substring(10);
                    typeof(CustomPrompt).GetProperty(field).SetValue(removed, field == "Id" ? Second : "different");
                }
                if (name == "identity-ignored") removed.Identity = "other-metadata";
                return ListCase(name, new List<CustomPrompt> { a, Prompt(), other }, removed, other,
                    (left, right) => left.Id == right.Id && left.Name == right.Name && left.AudioPrompt == right.AudioPrompt && left.TextPrompt == right.TextPrompt && left.Type == right.Type && left.Locale == right.Locale);
            }
            if (kind == "package") return ListCase(name, new List<CustomPromptPackage> { Package(), Package(), Package("b") }, Package(), Package("b"),
                (left, right) => { throw new ArgumentException("Cannot remove elements of Packages"); });
            throw new ArgumentException("Unknown list kind");
        }

        static string ListCase<T>(string name, List<T> values, T removed, T other, Func<T,T,bool> equal) where T : class
        {
            string failure = null;
            try {
                if (name == "remove-from-empty") values.Clear();
                if (name == "null-existing") values.Insert(0, null);
                PSListModifier<T> edit = name switch {
                    "null" => null,
                    "empty-operators" => new PSListModifier<T>(new Hashtable { { "Add", Array.Empty<T>() }, { "Remove", Array.Empty<T>() } }),
                    "replace-duplicates" => new PSListModifier<T>(new object[] { other, other, removed }),
                    "replace-null" => new PSListModifier<T>(new object[] { null }),
                    "remove-two" => new PSListModifier<T>(new Hashtable { { "Remove", new[] { removed, removed } } }),
                    "add-existing" => new PSListModifier<T>(new Hashtable { { "Add", new[] { removed, removed } } }),
                    "add-null" => new PSListModifier<T>(new Hashtable { { "Add", new T[] { null } } }),
                    "remove-null" => new PSListModifier<T>(new Hashtable { { "Remove", new T[] { null } } }),
                    "remove-and-add" => new PSListModifier<T>(new Hashtable { { "Remove", new[] { removed } }, { "Add", new[] { other } } }),
                    _ => new PSListModifier<T>(new Hashtable { { "Remove", new[] { removed } } })
                };
                edit.ApplyTo(values, equal);
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, values });
        }
    }
}
