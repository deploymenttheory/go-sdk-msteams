// Actual inherited compiled binding, with original lifecycle replaced offline.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Runspaces;
using System.Net.Http;
using System.Reflection;
using Newtonsoft.Json;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;

namespace TeamsContractOracle
{
    [Cmdlet("Invoke","EducationConfigurationGetFixture",DefaultParameterSetName="Identity")]
    public sealed class EducationConfigurationGetBinding : GetTeamsEducationConfigurationCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(JsonConvert.SerializeObject(new{parameter_set=ParameterSetName,bound=MyInvocation.BoundParameters}));}
        public string[] ApplyFilter(string pattern){Filter=pattern;return FilterPolicies(new List<TeamsEducationConfiguration>{new(){Identity="global"},new(){Identity="Tag:Fixture"},new(){Identity="Global"},new(){Identity=" "},new(){Identity="Global/fixture"}}).Select(v=>v.Identity).ToArray();}
    }
    [Cmdlet("Invoke","EducationConfigurationSetFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class EducationConfigurationSetBinding : SetTeamsEducationConfigurationCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(JsonConvert.SerializeObject(new{parameter_set=ParameterSetName,bound=MyInvocation.BoundParameters,specified=AreParametersSpecified(),settings=new{ParentGuardianPreferredContactMethod,EduGenerativeAIEnhancements,UpdateParentInformation}}));}
    }
    public static class EducationConfigurationBindingOracle {
        public static string Run(string verb,string arguments){
            var field=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);
            var old=field.GetValue(null);var oldHost=HostSingletionInstance.HostContract;
            var host=new ProviderFixtureHost();using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);field.SetValue(null,client);HostSingletionInstance.HostContract=host;
            string output=null,failure=null;bool hadErrors=false;
            try{
                var state=InitialSessionState.CreateDefault();
                var type=verb=="Get"?typeof(EducationConfigurationGetBinding):verb=="Set"?typeof(EducationConfigurationSetBinding):throw new ArgumentException("Unknown fixture verb");
                string name="Invoke-EducationConfiguration"+verb+"Fixture";state.Commands.Add(new SessionStateCmdletEntry(name,type,null));
                using var runspace=RunspaceFactory.CreateRunspace(state);runspace.Open();using var ps=PowerShell.Create();ps.Runspace=runspace;
                ps.AddCommand(name);foreach(var p in JsonConvert.DeserializeObject<Dictionary<string,object>>(arguments))ps.AddParameter(p.Key,p.Value);
                var values=ps.Invoke();hadErrors=ps.HadErrors;
                if(hadErrors)failure=ps.Streams.Error.First().Exception.GetType().FullName;
                if(values.Count==1)output=(string)values[0].BaseObject;
                if(!ReferenceEquals(field.GetValue(null),client))throw new InvalidOperationException("Constructor replaced memory client");
            }catch(Exception error){failure=(error.InnerException??error).GetType().FullName;}
            finally{field.SetValue(null,old);HostSingletionInstance.HostContract=oldHost;}
            return JsonConvert.SerializeObject(new{failure,had_errors=hadErrors,output,sends=sender.Capture.Sends,tokens=host.TokenRequests});
        }
        public static string Filter(string pattern){
            var field=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);var old=field.GetValue(null);var oldHost=HostSingletionInstance.HostContract;var host=new ProviderFixtureHost();
            using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);field.SetValue(null,client);HostSingletionInstance.HostContract=host;
            try{return JsonConvert.SerializeObject(new{identities=new EducationConfigurationGetBinding().ApplyFilter(pattern),sends=sender.Capture.Sends,tokens=host.TokenRequests});}
            finally{field.SetValue(null,old);HostSingletionInstance.HostContract=oldHost;}
        }
    }
}
