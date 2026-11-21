// Offline request-helper adapter. No cmdlet ProcessRecord or credential flow runs.
using System;
using System.Collections.Generic;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core.Custom.GPA;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core.Custom.GPA.Models;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;
using Newtonsoft.Json;

namespace TeamsContractOracle
{
    public static class AssignmentMigrationOracle
    {
        public static ProviderCapture Run(string operation, int status, string response, bool allow)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment())
                throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null);
            var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract;
            var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status, response);
            using var client = new HttpClient(sender);
            // Source-backed factory default; the factory cannot create a network sender.
            client.DefaultRequestHeaders.Add("MPACmdlet", "true");
            clientField.SetValue(null, client);
            versionField.SetValue(null, "8.1.0");
            HostSingletionInstance.HostContract = host;
            try
            {
                string command = operation switch {
                    "types" => "Get-GroupAssignmentRecommendationsPerPolicyType",
                    "names" => "Get-GroupAssignmentRecommendationsPerPolicyName",
                    "conflicts" => "Get-GroupPolicyAssignmentConflict",
                    "create" => "Invoke-StartDirectToGroupAssignmentMigration",
                    "clear" => "Invoke-ClearDirectToGroupAssignmentMigration",
                    _ => "Get-DirectToGroupAssignmentsMigrationStatus"
                };
                // CmdletInfo is metadata only. The command is never instantiated/invoked.
                var ctor = typeof(InvocationInfo).GetConstructor(BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public,
                    null, new[] { typeof(CommandInfo), typeof(IScriptExtent) }, null);
                var invocation = (InvocationInfo)ctor.Invoke(new object[] { new CmdletInfo(command, typeof(GetDirectToGroupAssignmentsMigrationStatus)), null });
                string suffix = operation switch {
                    "list" => "directAssignmentsMigration",
                    "read" => "directAssignmentsMigration/fixture-job",
                    "types" => "groups/recommendations?groupThreshold=500",
                    "names" => "groups/policyTypes/FixturePolicy/recommendations?groupThreshold=500",
                    "conflicts" => "fixture-group/policyTypes/FixturePolicy/conflict",
                    "create" => "directAssignmentsMigration/migrate?allowNonEmptyGroup=" + allow,
                    "clear" => "directAssignmentsMigration/cleanup?allowNonEmptyGroup=" + allow,
                    _ => throw new ArgumentException("Unknown synthetic operation")
                };
                string body = "";
                if (operation == "create" || operation == "clear") {
                    var model = new DirectToGroupAssignmentMigration {
                        Authority = "Tenant", PolicyType = "FixturePolicy", PolicyName = "Fixture",
                        GroupId = Guid.Parse("00000000-0000-4000-8000-000000000001")
                    };
                    body = JsonConvert.SerializeObject(model);
                }
                using var result = CommonHttpHelperRequest.SendGPARequests(body == "" ? "get" : "post",
                    "offline-fixture-only", "api.interfaces.records.teams.microsoft.com", invocation,
                    "00000000-0000-4000-8000-000000000003", suffix, body).GetAwaiter().GetResult();
                // This helper returns the HTTP response unchanged; command-level
                // deserialization/error-to-string behavior is separately source reviewed.
                sender.Capture.Success = result.IsSuccessStatusCode;
                if (result.StatusCode != (System.Net.HttpStatusCode)status || result.Content.ReadAsStringAsync().GetAwaiter().GetResult() != (response ?? ""))
                    throw new InvalidOperationException("Response pass-through differs");
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
    }
}
