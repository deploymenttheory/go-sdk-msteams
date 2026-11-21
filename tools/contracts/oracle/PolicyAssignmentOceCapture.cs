// Development-only adapter for the pinned OCE sender and model/configuration data.
// No InitializeServiceParameters, BeginProcessing or ProcessRecord is invoked.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Host;
using System.Net;
using System.Net.Http;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Teams.Policy.Administration.Cmdlets.OCE.Common;
using Microsoft.Teams.Policy.Administration.Cmdlets.OCE.Policies.GPA.Models;
using Microsoft.Teams.Policy.Administration.Cmdlets.OCE.Policies.LROS;
using Microsoft.Teams.Policy.Administration.Cmdlets.OCE.Policies.LROS.Models;
using Newtonsoft.Json;

namespace TeamsContractOracle
{
    public sealed class OceCapture
    {
        public string Method { get; set; }
        public string Url { get; set; }
        public string Body { get; set; }
        public bool TenantMatches { get; set; }
        public bool CorrelationIsGuid { get; set; }
        public bool HasMpa { get; set; }
        public bool HasCmdletName { get; set; }
        public bool HasTargetUri { get; set; }
        public int Sends { get; set; }
        public int? ReturnedStatus { get; set; }
        public string FailureType { get; set; }
        public int WarningCount { get; set; }
    }

    public sealed class OceRuntime : ICommandRuntime
    {
        public int Warnings;
        public PSHost Host => throw new InvalidOperationException("Offline runtime has no host");
        public PSTransactionContext CurrentPSTransaction => throw new InvalidOperationException("Offline runtime has no transaction");
        public void WriteDebug(string text) { }
        public void WriteError(ErrorRecord error) => throw new InvalidOperationException("Unexpected offline command error");
        public void WriteObject(object value) => throw new InvalidOperationException("Unexpected command processing");
        public void WriteObject(object value, bool enumerate) => WriteObject(value);
        public void WriteProgress(ProgressRecord value) { }
        public void WriteProgress(long id, ProgressRecord value) { }
        public void WriteVerbose(string text) { }
        public void WriteWarning(string text) { Warnings++; }
        public void WriteCommandDetail(string text) { }
        public bool ShouldProcess(string target) => throw new InvalidOperationException("No command approval in offline sender");
        public bool ShouldProcess(string target, string action) => ShouldProcess(target);
        public bool ShouldProcess(string description, string warning, string caption) => ShouldProcess(description);
        public bool ShouldProcess(string description, string warning, string caption, out ShouldProcessReason reason) { reason = ShouldProcessReason.None; return ShouldProcess(description); }
        public bool ShouldContinue(string query, string caption) => ShouldProcess(query);
        public bool ShouldContinue(string query, string caption, ref bool yes, ref bool no) => ShouldProcess(query);
        public bool TransactionAvailable() => false;
        public void ThrowTerminatingError(ErrorRecord error) => throw new InvalidOperationException("Unexpected offline terminating error");
    }

    // The base constructor only reads its packaged CmdletSettings.json. The
    // synthetic runtime consumes verbose/warning streams without printing them.
    public sealed class OfflineOce : IC3DirectoryTorusCmdletBase
    {
        public OfflineOce() { CommandRuntime = new OceRuntime(); }
        public string Suffix(string service, string operation)
        {
            Type type = typeof(IC3DirectoryTorusCmdletBase);
            Type serviceType = type.Assembly.GetType(type.Namespace + ".ServiceType", true);
            return (string)type.GetMethod("GetServiceUriSuffix", BindingFlags.Instance | BindingFlags.NonPublic).Invoke(this, new object[] { Enum.Parse(serviceType, service), operation });
        }
        public string Profiles()
        {
            var values = cmdletConfiguration.EnvironmentConfigurations.Select(c => new {
                environment = c.EnvironmentType.ToString(), origin = c.ConfigApiBaseUrl,
                regions = LROSConstants.EnvironmentToLrosRegionMapping[c.EnvironmentType].Select(r => r.ToString()).ToArray()
            }).ToArray();
            return JsonConvert.SerializeObject(values);
        }
    }

    public sealed class OceSender : HttpMessageHandler
    {
        readonly int status;
        readonly string reply;
        public readonly OceCapture Result = new OceCapture();
        public OceSender(int status, string reply) { this.status = status; this.reply = reply; }
        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken token)
        {
            if (request.RequestUri.Host != "api.interfaces.records.teams.microsoft.com" || request.Headers.Authorization?.Parameter != "offline-fixture-only")
                throw new InvalidOperationException("Unexpected offline destination or credential");
            Result.Sends++;
            Result.Method = request.Method.Method;
            Result.Url = request.RequestUri.AbsoluteUri;
            Result.Body = request.Content == null ? null : await request.Content.ReadAsStringAsync();
            Result.TenantMatches = request.Headers.GetValues("X-MS-Tenant-Id").Single() == "00000000-0000-4000-8000-000000000003";
            Result.CorrelationIsGuid = Guid.TryParse(request.Headers.GetValues("X-MS-Correlation-Id").Single(), out _);
            Result.HasMpa = request.Headers.Contains("MPACmdlet");
            Result.HasCmdletName = request.Headers.Contains("X-MS-CmdletName");
            Result.HasTargetUri = request.Headers.Contains("X-MS-Target-Uri");
            return new HttpResponseMessage((HttpStatusCode)status) { Content = new StringContent(reply ?? "") };
        }
    }

    public static class PolicyAssignmentOceOracle
    {
        public static OceCapture Run(string service, string operation, string method, Dictionary<string,string> replacements, int status, string reply)
        {
            var previousSettings = JsonConvert.DefaultSettings;
            FieldInfo field = typeof(RestClient).GetField("HttpClient", BindingFlags.Static | BindingFlags.NonPublic);
            object previousClient = field.GetValue(null);
            using var sender = new OceSender(status, reply);
            using var memoryClient = new HttpClient(sender);
            try
            {
                var cmdlet = new OfflineOce();
                string suffix = cmdlet.Suffix(service, operation);
                foreach (var item in replacements) suffix = suffix.Replace("{" + item.Key + "}", item.Value);
                if (suffix.Contains("{") || suffix.Contains("}")) throw new InvalidOperationException("Incomplete synthetic route");
                var uri = new Uri(new Uri("https://api.interfaces.records.teams.microsoft.com"), suffix);
                var headers = cmdlet.InitializeHeaders(Guid.Parse("00000000-0000-4000-8000-000000000003"));
                field.SetValue(null, memoryClient);
                try {
                    using var response = RestClient.SendHttpRequest(new HttpMethod(method), uri, "offline-fixture-only", headers, null, service, cmdlet);
                    sender.Result.ReturnedStatus = (int)response.StatusCode;
                }
                catch (Exception error) { sender.Result.FailureType = error.GetType().Name; }
                sender.Result.WarningCount = ((OceRuntime)cmdlet.CommandRuntime).Warnings;
                return sender.Result;
            }
            finally { field.SetValue(null, previousClient); JsonConvert.DefaultSettings = previousSettings; }
        }

        public static string Profiles()
        {
            var previous = JsonConvert.DefaultSettings;
            try { return new OfflineOce().Profiles(); }
            finally { JsonConvert.DefaultSettings = previous; }
        }

        public static string Project(string model, string input)
        {
            Type type = model switch {
                "GroupAssignment" => typeof(GroupPolicyAssignment),
                "GroupMembers" => typeof(GroupMembersResponse),
                "UserMembership" => typeof(UserMembershipResponse),
                "BatchDefinition" => typeof(BatchOperationDefinition),
                "BatchStatus" => typeof(BatchOperationStatus),
                "ExtendedBatchStatus" => typeof(ExtendedBatchOperationStatus),
                _ => throw new ArgumentException("Unknown OCE model fixture")
            };
            return JsonConvert.SerializeObject(JsonConvert.DeserializeObject(input, type));
        }
    }
}
