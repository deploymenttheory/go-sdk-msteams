// Offline adapter for pure pinned domain normalization. No command processor,
// connection, token source, HTTP helper or tenant state is used.
using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using Microsoft.Teams.ConfigApi.Cmdlets;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models;

namespace TeamsContractOracle
{
    public static class FederationDomains
    {
        static PSObject Choice(bool all, string[] domains)
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
            return choice;
        }

        public static string Normalize(bool all, string[] domains)
        {
            var bag = new Dictionary<string, object>();
            bag.Add("Identity", "Global");
            bag.Add("AllowedDomains", Choice(all, domains));
            return NormalizeBag(bag);
        }

        static PSObject DomainObject()
        {
            var value = new PSObject();
            value.Properties.Add(new PSNoteProperty("Domain", "object.example.invalid"));
            value.Properties.Add(new PSNoteProperty("Enabled", false));
            value.Properties.Add(new PSNoteProperty("DataSource", "fixture-private"));
            value.Properties.Add(new PSNoteProperty("Key", "fixture-key"));
            return value;
        }

        public static string NormalizeCase(string name)
        {
            var bag = new Dictionary<string, object>();
            bag.Add("Identity", "Global");
            switch (name)
            {
                case "omitted": break;
                case "bound-scalars":
                    foreach (var key in new[] { "AllowFederatedUsers", "AllowTeamsConsumer", "AllowTeamsConsumerInbound", "RestrictTeamsConsumerToExternalUserProfiles", "SharedSipAddressSpace", "TreatDiscoveredPartnersAsUnverified", "BlockAllSubdomains", "EnableExternalAccessRestrictionsForChatParticipants", "EnableMutualFederationForChatParticipants" })
                        bag.Add(key, false);
                    bag.Add("ExternalAccessWithTrialTenants", "FutureState");
                    bag.Add("SecurityTeamAllowBlockListDelegation", "");
                    break;
                case "allow-all": bag.Add("AllowedDomains", Choice(true, Array.Empty<string>())); break;
                case "allow-list": bag.Add("AllowedDomains", Choice(false, new[] { "b.example.invalid", "a.example.invalid", "b.example.invalid", "" })); break;
                case "empty-allow-list": bag.Add("AllowedDomains", Choice(false, Array.Empty<string>())); break;
                case "allowed-replacement": bag.Add("AllowedDomainsAsAList", new[] { "a.example.invalid", "b.example.invalid" }); break;
                case "allowed-operators": bag.Add("AllowedDomainsAsAList", new Hashtable { { "Add", new[] { "a.example.invalid" } }, { "Remove", new[] { "b.example.invalid" } } }); break;
                case "allowed-empty-operators": bag.Add("AllowedDomainsAsAList", new Hashtable { { "Add", Array.Empty<string>() }, { "Remove", Array.Empty<string>() } }); break;
                case "blocked-mixed": bag.Add("BlockedDomains", new object[] { "a.example.invalid", DomainObject() }); break;
                case "blocked-operators": bag.Add("BlockedDomains", new Hashtable { { "Add", new object[] { DomainObject(), "a.example.invalid" } }, { "Remove", new[] { "b.example.invalid" } } }); break;
                case "trial-mixed": bag.Add("AllowedTrialTenantDomains", new object[] { DomainObject(), "a.example.invalid" }); break;
                case "trial-null": bag.Add("AllowedTrialTenantDomains", null); break;
                case "bound-null-scalar": bag.Add("AllowFederatedUsers", null); break;
                case "conflicting-choices": bag.Add("AllowedDomains", Choice(true, Array.Empty<string>())); bag.Add("AllowedDomainsAsAList", Array.Empty<string>()); break;
                case "null-choice": bag.Add("AllowedDomains", null); break;
                case "replace-operator": bag.Add("BlockedDomains", new Hashtable { { "Replace", new[] { "a.example.invalid" } } }); break;
                default: throw new ArgumentException("Unknown synthetic case");
            }
            return NormalizeBag(bag);
        }

        static string NormalizeBag(Dictionary<string, object> bag)
        {
            var cmdlet = new SetCsTenantFederationConfiguration();
            cmdlet.ValidateAndUpdatePropertyBagParams(bag);
            return ((XdsConfiguration)XdsConfiguration.DeserializeFromDictionary(bag)).ToJsonString();
        }
    }
}
