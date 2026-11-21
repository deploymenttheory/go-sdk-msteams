// Development-only pinned generated model projection. No cmdlets or I/O.
using System;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models;

namespace TeamsContractOracle
{
    public static class PolicyAssignmentModelOracle
    {
        public static string Project(string name, string input)
        {
            Type type = name switch {
                "GroupAssignment" => typeof(GroupAssignment),
                "EffectivePolicy" => typeof(EffectivePolicy),
                "SourceEntry" => typeof(SourceEntry),
                "BatchJobStatus" => typeof(BatchJobStatus),
                "SimpleBatchJobStatus" => typeof(SimpleBatchJobStatus),
                "GrantActionStatus" => typeof(GrantActionStatus),
                _ => throw new ArgumentException("Unknown model fixture")
            };
            object value = type.GetMethod("FromJsonString", BindingFlags.Public | BindingFlags.Static).Invoke(null, new object[] { input });
            return value == null ? "null" : (string)type.GetMethod("ToJsonString").Invoke(value, null);
        }
    }
}
