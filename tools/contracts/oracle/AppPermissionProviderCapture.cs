// Development-only pinned provider/model observations; never a command lifecycle.
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
    public static class AppPermissionOracle
    {
        public static ProviderCapture Run(string operation, string fields, int status, string response, bool force, string policyName, bool nullPolicyName)
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
                // The PowerShell method binder converts a null string argument
                // to empty; select true null inside C# for the helper contract.
                if (nullPolicyName) policyName = null;
                const string identity = "scope /Fixture / Child", token = "offline-fixture-only", origin = "api.interfaces.records.teams.microsoft.com";
                string command = operation switch {
                    "create" => "New-CsTeamsAppPermissionPolicy", "patch" => "Set-CsTeamsAppPermissionPolicy", "delete" => "Remove-CsTeamsAppPermissionPolicy",
                    "user" or "tenant" or "group" => "Grant-CsTeamsAppPermissionPolicy", _ => "Get-CsTeamsAppPermissionPolicy"
                };
                var bound = new Dictionary<string, object> { { "Identity", identity } };
                if (force) bound.Add("Force", new SwitchParameter(false));
                var value = new TeamsAppPermissionPolicy { Identity = "read-model-identity", Description = "", DefaultCatalogAppsType = "FutureDefaultMode", GlobalCatalogAppsType = "", PrivateCatalogAppsType = null };
                if (fields == "all") {
                    value.DefaultCatalogApps.Add(new DefaultCatalogApp("fixture-default"));
                    value.GlobalCatalogApps.Add(new GlobalCatalogApp("fixture-global"));
                    value.PrivateCatalogApps.Add(new PrivateCatalogApp("fixture-private"));
                }
                if (operation == "create" || operation == "patch") {
                    if (fields == "all" || fields == "description") {
                        foreach (var property in typeof(TeamsAppPermissionPolicy).GetProperties().Where(p => p.PropertyType == typeof(string) && p.Name != "Identity")) bound.Add(property.Name, property.GetValue(value));
                    }
                    if (fields == "all" || fields == "empty-lists") {
                        bound.Add("DefaultCatalogApps", new PSListModifier<DefaultCatalogApp>(value.DefaultCatalogApps.ToArray()));
                        bound.Add("GlobalCatalogApps", new PSListModifier<GlobalCatalogApp>(value.GlobalCatalogApps.ToArray()));
                        bound.Add("PrivateCatalogApps", new PSListModifier<PrivateCatalogApp>(value.PrivateCatalogApps.ToArray()));
                    }
                }
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetTeamsAppPermissionPolicyCmdlet)), null });
                foreach (var item in bound) invocation.BoundParameters.Add(item.Key, item.Value);
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<TeamsAppPermissionPolicy>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<TeamsAppPermissionPolicy>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<TeamsAppPermissionPolicy>.CreateAsync(value, identity, token, origin, invocation, null).GetAwaiter().GetResult(),
                    "patch" => HttpRequestHelper<TeamsAppPermissionPolicy>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<TeamsAppPermissionPolicy>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "user" => HttpRequestHelper<TeamsAppPermissionPolicy>.GrantToUserAsync("user/fixture", policyName, token, origin, invocation).GetAwaiter().GetResult(),
                    "tenant" => HttpRequestHelper<TeamsAppPermissionPolicy>.GrantToTenantAsync(policyName, token, origin, invocation, false).GetAwaiter().GetResult(),
                    "group" => HttpRequestHelper<TeamsAppPermissionPolicy>.GrantToGroupAsync("group/fixture", policyName, token, origin, invocation, -1).GetAwaiter().GetResult(),
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

        public static string PolicyModel(string input)
        {
            object value = null; string failure = null;
            try {
                value = input switch {
                    "constructor" => new TeamsAppPermissionPolicy(),
                    "future-types-long-description" => new TeamsAppPermissionPolicy { Description = new string('x', 2000), DefaultCatalogAppsType = "FutureMode", GlobalCatalogAppsType = "", PrivateCatalogAppsType = null },
                    _ => JsonConvert.DeserializeObject<TeamsAppPermissionPolicy>(input)
                };
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, value });
        }

        public static string AppModel(string catalog, string input)
        {
            object value = null; string failure = null;
            try {
                Type type = catalog switch { "default" => typeof(DefaultCatalogApp), "global" => typeof(GlobalCatalogApp), "private" => typeof(PrivateCatalogApp), _ => throw new ArgumentException("Unknown catalog") };
                if (input.StartsWith("json:")) value = JsonConvert.DeserializeObject(input.Substring(5), type);
                else value = Activator.CreateInstance(type, new object[] { input == "null" ? null : input });
            } catch (Exception error) { failure = (error.InnerException ?? error).GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, value });
        }

        public static string Lists(string catalog, string name)
        {
            return catalog switch {
                "default" => ListCase(name, id => new DefaultCatalogApp(id), item => item.Id),
                "global" => ListCase(name, id => new GlobalCatalogApp(id), item => item.Id),
                "private" => ListCase(name, id => new PrivateCatalogApp(id), item => item.Id),
                _ => throw new ArgumentException("Unknown catalog")
            };
        }
        static string ListCase<T>(string name, Func<string,T> create, Func<T,string> id)
        {
            var a = create("a"); var duplicate = create("a"); var upper = create("A"); var b = create("b");
            var values = new List<T> { a, duplicate, upper, b }; string failure = null;
            try {
                PSListModifier<T> edit = name switch {
                    "null" => null,
                    "empty-operators" => new PSListModifier<T>(new Hashtable { { "Add", Array.Empty<T>() }, { "Remove", Array.Empty<T>() } }),
                    "replace-duplicates" => new PSListModifier<T>(new object[] { b, b, a }),
                    "remove-one" => new PSListModifier<T>(new Hashtable { { "Remove", new[] { duplicate } } }),
                    "remove-two" => new PSListModifier<T>(new Hashtable { { "Remove", new[] { duplicate, duplicate } } }),
                    "add-existing" => new PSListModifier<T>(new Hashtable { { "Add", new[] { duplicate, duplicate } } }),
                    "remove-and-add" => new PSListModifier<T>(new Hashtable { { "Remove", new[] { duplicate } }, { "Add", new[] { duplicate } } }),
                    _ => throw new ArgumentException("Unknown synthetic edit")
                };
                edit.ApplyTo(values, (left,right) => id(left) == id(right));
            } catch (Exception error) { failure = error.GetType().FullName; }
            return JsonConvert.SerializeObject(new { failure, values });
        }
    }
}
