// Offline adapter for pure pinned domain normalization. No command processor,
// connection, token source, HTTP helper or tenant state is used.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using Microsoft.Teams.ConfigApi.Cmdlets;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models;

namespace TeamsContractOracle
{
    public static class FederationDomains
    {
        public static string Normalize(bool all, string[] domains)
        {
            PSObject choice;
            if (all)
            {
                choice = PSObject.AsPSObject(new NewCsEdgeAllowAllKnownDomainsHelper.AllowAllKnownDomains());
            }
            else
            {
                var values = domains.Select(NewEdgeDomainPatternHelper.ConvertToDomainPattern).ToArray();
                choice = new PSObject();
                choice.Properties.Add(new PSNoteProperty("AllowedDomain", PSObject.AsPSObject(values)));
            }
            var bag = new Dictionary<string, object>();
            bag.Add("Identity", "Global");
            bag.Add("AllowedDomains", choice);
            var cmdlet = new SetCsTenantFederationConfiguration();
            cmdlet.ValidateAndUpdatePropertyBagParams(bag);
            return ((XdsConfiguration)XdsConfiguration.DeserializeFromDictionary(bag)).ToJsonString();
        }
    }
}
