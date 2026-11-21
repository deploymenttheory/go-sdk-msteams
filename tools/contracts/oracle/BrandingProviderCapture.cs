// Offline direct provider calls and local SMA list behavior. No cmdlet processing.
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
    public static class BrandingOracle
    {
        static InvocationInfo Invocation(string name, Dictionary<string, object> bound)
        {
            var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
            var result = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(name, typeof(GetTeamsMeetingBrandingPolicyCmdlet)), null });
            foreach (var item in bound) result.BoundParameters.Add(item.Key, item.Value);
            return result;
        }

        public static ProviderCapture Run(string operation, int status, string response, bool force, bool nested)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null); var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract;
            var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status, response);
            using var client = new HttpClient(sender);
            client.DefaultRequestHeaders.Add("MPACmdlet", "true");
            clientField.SetValue(null, client); versionField.SetValue(null, "8.1.0"); HostSingletionInstance.HostContract = host;
            try
            {
                const string identity = "scope /Fixture / Child", token = "offline-fixture-only", origin = "api.interfaces.records.teams.microsoft.com";
                string command = operation switch {
                    "create" => "New-CsTeamsMeetingBrandingPolicy", "patch" => "Set-CsTeamsMeetingBrandingPolicy", "delete" => "Remove-CsTeamsMeetingBrandingPolicy",
                    "user" => "Grant-CsTeamsMeetingBrandingPolicy", "tenant" => "Grant-CsTeamsMeetingBrandingPolicy", "group" => "Grant-CsTeamsMeetingBrandingPolicy",
                    _ => "Get-CsTeamsMeetingBrandingPolicy"
                };
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (force) bound.Add("Force", new SwitchParameter(false));
                var policy = new TeamsMeetingBrandingPolicy { Identity = "read-model-identity", DefaultTheme = "", EnableMeetingOptionsThemeOverride = false,
                    EnableMeetingOptionsBrandedReactionsOverride = false, EnableNdiAssuranceSlate = false, EnableMeetingBackgroundImages = false, RequireBackgroundEffect = false };
                if (nested) {
                    policy.MeetingBackgroundImages.Add(new MeetingBackgroundImage("fixture-image", 0, "Fixture", "https://example.invalid/background.png") { IsRequired = false });
                    policy.MeetingBrandingThemes.Add(new MeetingBrandingTheme("fixture-theme") { Enabled = false });
                    policy.NdiAssuranceSlateImages.Add(new NdiAssuranceSlate("fixture-slate", "Fixture", "https://example.invalid/slate.png") { IsDefault = false });
                }
                if (operation == "create" || operation == "patch") {
                    foreach (var key in new[] { "DefaultTheme", "EnableMeetingOptionsThemeOverride", "EnableMeetingOptionsBrandedReactionsOverride", "EnableNdiAssuranceSlate", "EnableMeetingBackgroundImages", "RequireBackgroundEffect", "MeetingBackgroundImages", "MeetingBrandingThemes", "NdiAssuranceSlateImages" })
                        bound.Add(key, typeof(TeamsMeetingBrandingPolicy).GetProperty(key).GetValue(policy));
                }
                var invocation = Invocation(command, bound);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.CreateAsync(policy, identity, token, origin, invocation, null).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.SetAsync(policy, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "user" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.GrantToUserAsync("user/fixture", null, token, origin, invocation).GetAwaiter().GetResult(),
                    "tenant" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.GrantToTenantAsync("", token, origin, invocation, false).GetAwaiter().GetResult(),
                    "group" => HttpRequestHelper<TeamsMeetingBrandingPolicy>.GrantToGroupAsync("group/fixture", null, token, origin, invocation, -1).GetAwaiter().GetResult(),
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

        public static string ApplyList(string kind, string name, bool empty)
        {
            // Delegates are exactly those source-reviewed in the pinned cmdlets.
            bool rejected = false;
            string[] ids;
            if (kind == "policy") {
                var a = new MeetingBackgroundImage("a", 0, "Fixture", "https://example.invalid/a");
                var b = new MeetingBackgroundImage("b", 1, "Fixture", "https://example.invalid/b");
                var values = empty ? new List<MeetingBackgroundImage>() : new List<MeetingBackgroundImage> { a, a, b };
                PSListModifier<MeetingBackgroundImage> modifier = name switch {
                    "null" => null,
                    "empty-replace" => new PSListModifier<MeetingBackgroundImage>(Array.Empty<object>()),
                    "replace" => new PSListModifier<MeetingBackgroundImage>(new object[] { b, b }),
                    "empty-delta" => new PSListModifier<MeetingBackgroundImage>(new Hashtable { { "Add", Array.Empty<MeetingBackgroundImage>() }, { "Remove", Array.Empty<MeetingBackgroundImage>() } }),
                    "add" => new PSListModifier<MeetingBackgroundImage>(new Hashtable { { "Add", new[] { a } } }),
                    "remove" => new PSListModifier<MeetingBackgroundImage>(new Hashtable { { "Remove", new[] { a } } }),
                    "remove-missing" => new PSListModifier<MeetingBackgroundImage>(new Hashtable { { "Remove", new[] { new MeetingBackgroundImage("missing", 0, "Fixture", "https://example.invalid/a") } } }),
                    "null-item" => new PSListModifier<MeetingBackgroundImage>(new object[] { null }),
                    _ => throw new ArgumentException("Unknown list case")
                };
                try { modifier.ApplyTo(values, (x, y) => throw new ArgumentException("Source comparator forbids removal")); }
                catch { rejected = true; }
                ids = values.Select(v => v?.Id).ToArray();
            } else {
                var a = new MeetingReactionItemModel("a", "https://example.invalid/a", "Like");
                var different = new MeetingReactionItemModel("a", "https://example.invalid/a", "Like") { BrandedDisplayName = "different" };
                var b = new MeetingReactionItemModel("b", "https://example.invalid/b", "Future");
                var values = empty ? new List<MeetingReactionItemModel>() : new List<MeetingReactionItemModel> { a, a, different };
                PSListModifier<MeetingReactionItemModel> modifier = name switch {
                    "null" => null,
                    "empty-replace" => new PSListModifier<MeetingReactionItemModel>(Array.Empty<object>()),
                    "replace" => new PSListModifier<MeetingReactionItemModel>(new object[] { b, b }),
                    "empty-delta" => new PSListModifier<MeetingReactionItemModel>(new Hashtable { { "Add", Array.Empty<MeetingReactionItemModel>() }, { "Remove", Array.Empty<MeetingReactionItemModel>() } }),
                    "add" => new PSListModifier<MeetingReactionItemModel>(new Hashtable { { "Add", new[] { a } } }),
                    "remove" => new PSListModifier<MeetingReactionItemModel>(new Hashtable { { "Remove", new[] { a } } }),
                    "remove-different" => new PSListModifier<MeetingReactionItemModel>(new Hashtable { { "Remove", new[] { different } } }),
                    _ => throw new ArgumentException("Unknown reaction case")
                };
                try { modifier.ApplyTo(values, (x, y) => x.Id == y.Id && x.ReactionImageUrl == y.ReactionImageUrl && x.ReactionSpriteImageUrl == y.ReactionSpriteImageUrl && x.ReactionType == y.ReactionType && x.BrandedDisplayName == y.BrandedDisplayName); }
                catch { rejected = true; }
                ids = values.Select(v => v?.Id + (v?.BrandedDisplayName == null ? "" : ":" + v.BrandedDisplayName)).ToArray();
            }
            return JsonConvert.SerializeObject(new { rejected, ids });
        }
    }
}
