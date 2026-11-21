// Offline-only adapter for pinned compiled provider request/response methods.
// The host refuses token acquisition, and the only HTTP sender is memory-only.
using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net;
using System.Net.Http;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;

namespace TeamsContractOracle
{
    public sealed class ProviderCapture
    {
        public string Method { get; set; }
        public string Url { get; set; }
        public string Body { get; set; }
        public string ContentType { get; set; }
        public string CmdletName { get; set; }
        public bool CorrelationIsGuid { get; set; }
        public bool HasTargetUri { get; set; }
        public string MpaCmdlet { get; set; }
        public int Sends { get; set; }
        public int TokenRequests { get; set; }
        public bool? Success { get; set; }
        public bool? DecodedNull { get; set; }
        public string DecodedType { get; set; }
        public string FailureType { get; set; }
        public string[] Warnings { get; set; }
    }

    public sealed class ProviderFixtureHost : ICmdletHostContract
    {
        public ILogMessage Logger { get; set; }
        public IHostTelemetryClient TelemetryClient { get; set; }
        public int TokenRequests;
        public LoginInfo GetAccessToken(TokenGenerationInfo info)
        {
            TokenRequests++;
            throw new InvalidOperationException("Offline host refuses credentials");
        }
        public void GetModulesInfo(out string config, out string teams)
        {
            config = "8.1.0";
            teams = "8.1.0";
        }
    }

    public sealed class ProviderSender : HttpMessageHandler
    {
        readonly int status;
        readonly string response;
        public readonly ProviderCapture Capture = new ProviderCapture();
        public ProviderSender(int status, string response) { this.status = status; this.response = response; }
        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken token)
        {
            if (request.RequestUri.Host != "api.interfaces.records.teams.microsoft.com" || request.Headers.Authorization?.Parameter != "offline-fixture-only")
                throw new InvalidOperationException("Unexpected offline destination or token");
            Capture.Sends++;
            Capture.Method = request.Method.Method;
            Capture.Url = request.RequestUri.AbsoluteUri;
            Capture.Body = request.Content == null ? null : await request.Content.ReadAsStringAsync();
            Capture.ContentType = request.Content?.Headers.ContentType?.ToString();
            Capture.CmdletName = request.Headers.GetValues("X-MS-CmdletName").Single();
            Capture.CorrelationIsGuid = Guid.TryParse(request.Headers.GetValues("X-MS-Correlation-Id").Single(), out _);
            Capture.HasTargetUri = request.Headers.Contains("X-MS-Target-Uri");
            Capture.MpaCmdlet = request.Headers.GetValues("MPACmdlet").Single();
            var result = new HttpResponseMessage((HttpStatusCode)status) { Content = new StringContent(response ?? "") };
            result.Headers.Add("X-Ms-ConfigApi-PowerShell-WarningMessage", "fixture-warning");
            return result;
        }
    }

    public static class ProviderOracle
    {
        public static string ApplyListCase(string name)
        {
            var values = new List<string> { "a", "a", "A", "b" };
            PSListModifier<string> modifier = name switch {
                "null" => null,
                "empty-operators" => new PSListModifier<string>(new Hashtable { { "Add", Array.Empty<string>() }, { "Remove", Array.Empty<string>() } }),
                "replacement" => new PSListModifier<string>(new object[] { "b", "b", "" }),
                "remove-one" => new PSListModifier<string>(new Hashtable { { "Remove", new[] { "a" } } }),
                "remove-missing" => new PSListModifier<string>(new Hashtable { { "Remove", new[] { "missing" } } }),
                "add-existing" => new PSListModifier<string>(new Hashtable { { "Add", new[] { "a", "c", "c" } } }),
                "remove-and-add" => new PSListModifier<string>(new Hashtable { { "Remove", new[] { "a", "b" } }, { "Add", new[] { "b", "a" } } }),
                _ => throw new ArgumentException("Unknown synthetic list case")
            };
            bool empty = modifier == null || modifier.IsEmpty();
            if (empty) values.Clear(); else modifier.ApplyTo(values);
            return System.Text.Json.JsonSerializer.Serialize(new { is_empty = empty, values });
        }

        static InvocationInfo Invocation(string name, Dictionary<string, object> bound)
        {
            var constructor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
            var value = (InvocationInfo)constructor.Invoke(new object[] { new CmdletInfo(name, typeof(GetExternalAccessPolicyCmdlet)), null });
            foreach (var item in bound) value.BoundParameters.Add(item.Key, item.Value);
            return value;
        }

        public static ProviderCapture Run(string operation, int status, string response, bool forceBound)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null);
            var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract;
            var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status, response);
            using var client = new HttpClient(sender);
            // Mirrors the single default header of CreateHttpClient; handler
            // injection prevents that factory from creating a network sender.
            client.DefaultRequestHeaders.Add("MPACmdlet", "true");
            clientField.SetValue(null, client);
            versionField.SetValue(null, "8.1.0");
            HostSingletionInstance.HostContract = host;
            try
            {
                var bound = new Dictionary<string, object>();
                string identity = "scope /Fixture / Child";
                if(operation=="read-tagged") {operation="read";identity="tag:Fixture / Policy";}
                bound.Add("Identity", identity);
                if (forceBound) bound.Add("Force", new SwitchParameter(false));
                string command = operation switch {
                    "create" => "New-CsExternalAccessPolicy", "update" => "Set-CsExternalAccessPolicy", "delete" => "Remove-CsExternalAccessPolicy",
                    "user" => "Grant-CsExternalAccessPolicy", "tenant" => "Grant-CsExternalAccessPolicy", "group" => "Grant-CsExternalAccessPolicy",
                    "acs-read" => "Get-CsTeamsAcsFederationConfiguration", "acs-list" => "Get-CsTeamsAcsFederationConfiguration", "acs-update" => "Set-CsTeamsAcsFederationConfiguration",
                    "external-read" => "Get-CsTeamsExternalAccessConfiguration", "external-list" => "Get-CsTeamsExternalAccessConfiguration", "external-update" => "Set-CsTeamsExternalAccessConfiguration",
                    _ => "Get-CsExternalAccessPolicy"
                };
                var policy = new ExternalAccessPolicy { Identity = "read-model-identity", Description = "", EnableFederationAccess = false };
                if (operation == "create" || operation == "update") {
                    bound.Add("Description", ""); bound.Add("EnableFederationAccess", false);
                }
                var invocation = Invocation(command, bound);
                const string token = "offline-fixture-only", origin = "api.interfaces.records.teams.microsoft.com";
                PolicyResponse result = operation switch {
                    "read" => HttpRequestHelper<ExternalAccessPolicy>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "list" => HttpRequestHelper<ExternalAccessPolicy>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "create" => HttpRequestHelper<ExternalAccessPolicy>.CreateAsync(policy, identity, token, origin, invocation, null).GetAwaiter().GetResult(),
                    "update" => HttpRequestHelper<ExternalAccessPolicy>.SetAsync(policy, identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "delete" => HttpRequestHelper<ExternalAccessPolicy>.RemoveAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "user" => HttpRequestHelper<ExternalAccessPolicy>.GrantToUserAsync("user/fixture", null, token, origin, invocation).GetAwaiter().GetResult(),
                    "tenant" => HttpRequestHelper<ExternalAccessPolicy>.GrantToTenantAsync("", token, origin, invocation, false).GetAwaiter().GetResult(),
                    "group" => HttpRequestHelper<ExternalAccessPolicy>.GrantToGroupAsync("group/fixture", null, token, origin, invocation, -1).GetAwaiter().GetResult(),
                    "acs-read" => HttpRequestHelper<TeamsAcsFederationConfiguration>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "acs-list" => HttpRequestHelper<TeamsAcsFederationConfiguration>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "acs-update" => UpdateAcs(identity, token, origin, invocation),
                    "external-read" => HttpRequestHelper<TeamsExternalAccessConfiguration>.GetAsync(identity, token, origin, invocation).GetAwaiter().GetResult(),
                    "external-list" => HttpRequestHelper<TeamsExternalAccessConfiguration>.GetAllAsync(token, origin, invocation).GetAwaiter().GetResult(),
                    "external-update" => UpdateExternal(identity, token, origin, invocation),
                    _ => throw new ArgumentException("Unknown synthetic operation")
                };
                sender.Capture.Success = result.IsSuccessStatusCode;
                sender.Capture.DecodedNull = result.PolicyResult == null;
                sender.Capture.DecodedType = result.PolicyResult?.GetType().FullName;
                sender.Capture.Warnings = result.Warnings?.ToArray();
            }
            catch (Exception error) { sender.Capture.FailureType = error.GetType().FullName; }
            finally {
                sender.Capture.TokenRequests = host.TokenRequests;
                HostSingletionInstance.HostContract = oldHost;
                versionField.SetValue(null, oldVersion);
                clientField.SetValue(null, oldClient);
            }
            return sender.Capture;
        }

        static PolicyResponse UpdateAcs(string identity, string token, string origin, InvocationInfo invocation)
        {
            invocation.BoundParameters.Add("EnableAcsUsers", false);
            invocation.BoundParameters.Add("AllowedAcsResources", Array.Empty<string>());
            var value = new TeamsAcsFederationConfiguration { Identity = "read-model-identity", EnableAcsUsers = false };
            return HttpRequestHelper<TeamsAcsFederationConfiguration>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult();
        }
        static PolicyResponse UpdateExternal(string identity, string token, string origin, InvocationInfo invocation)
        {
            invocation.BoundParameters.Add("BlockExternalUserAccess", null);
            invocation.BoundParameters.Add("BlockedUsers", Array.Empty<string>());
            var value = new TeamsExternalAccessConfiguration { Identity = "read-model-identity", BlockExternalUserAccess = null };
            return HttpRequestHelper<TeamsExternalAccessConfiguration>.SetAsync(value, identity, token, origin, invocation).GetAwaiter().GetResult();
        }
    }
}
