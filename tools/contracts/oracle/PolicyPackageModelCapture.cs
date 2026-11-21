// Pinned generated data-model projection only. No cmdlets, sender or credentials.
using System;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models;

namespace TeamsContractOracle
{
    public static class PolicyPackageModelOracle
    {
        public static string Project(string name, string input)
        {
            Type type = name switch {
                "Package" => typeof(PackageServiceModelsFormattedPackage),
                "Summary" => typeof(PackageServiceModelsFormattedPackageSummary),
                "Recommendation" => typeof(PackageServiceModelsFormattedPackageRecommendation),
                "CustomResponse" => typeof(PackageServiceModelsCustomPackageResponse),
                "GroupResponse" => typeof(PackageServiceModelsApplyPackageGroupResponse),
                "UserResponse" => typeof(PackageServiceModelsPostPackageResponse),
                "BatchResponse" => typeof(PackageServiceModelsBatchPostPackageResponse),
                "PolicyPair" => typeof(PackageServiceModelsRequestsPolicyTypeAndName),
                "Ranking" => typeof(PackageServiceModelsRequestsPolicyRanking),
                _ => throw new ArgumentException("Unknown package model fixture")
            };
            object value = type.GetMethod("FromJsonString", BindingFlags.Public | BindingFlags.Static).Invoke(null, new object[] { input });
            return value == null ? "null" : (string)type.GetMethod("ToJsonString").Invoke(value, null);
        }
    }
}
