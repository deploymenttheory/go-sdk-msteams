// Inherited compiled binder/pure helpers; original command lifecycle is replaced.
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
    static class FeedbackBound {
        public static string Value(PSCmdlet cmdlet,InvocationInfo info,string set,bool? specified=null) {
            var settings=new Dictionary<string,object>();
            foreach(var p in typeof(TeamsFeedbackPolicy).GetProperties().Where(p=>p.Name!="Identity")) {
                var source=cmdlet.GetType().GetProperty(p.Name);
                if(source!=null)settings.Add(p.Name,source.GetValue(cmdlet));
            }
            return JsonConvert.SerializeObject(new{parameter_set=set,bound=info.BoundParameters,specified,settings});
        }
    }
    [Cmdlet("Invoke","FeedbackGetFixture",DefaultParameterSetName="Identity")]
    public sealed class FeedbackGetBinding : GetTeamsFeedbackPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(FeedbackBound.Value(this,MyInvocation,ParameterSetName));}
        public string[] ApplyFilter(string pattern){Filter=pattern;return FilterPolicies(new List<TeamsFeedbackPolicy>{new(){Identity="global"},new(){Identity="Tag:Fixture"},new(){Identity="Global"},new(){Identity=" "},new(){Identity="Global/fixture"}}).Select(p=>p.Identity).ToArray();}
    }
    [Cmdlet("Invoke","FeedbackNewFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class FeedbackNewBinding : NewTeamsFeedbackPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(FeedbackBound.Value(this,MyInvocation,ParameterSetName,AreParametersSpecified()));}
    }
    [Cmdlet("Invoke","FeedbackSetFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class FeedbackSetBinding : SetTeamsFeedbackPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(FeedbackBound.Value(this,MyInvocation,ParameterSetName,AreParametersSpecified()));}
    }
    [Cmdlet("Invoke","FeedbackRemoveFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class FeedbackRemoveBinding : RemoveTeamsFeedbackPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(FeedbackBound.Value(this,MyInvocation,ParameterSetName));}
    }
    public static class FeedbackBindingOracle {
        public static string Run(string verb,string arguments){
            if(PowershellUtils.IsSelfHost()||PowershellUtils.IsTorusEnvironment())throw new InvalidOperationException("Public offline profile required");
            var field=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);
            var old=field.GetValue(null);var oldHost=HostSingletionInstance.HostContract;var host=new ProviderFixtureHost();
            using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);field.SetValue(null,client);HostSingletionInstance.HostContract=host;
            string output=null,failure=null;bool hadErrors=false;
            try{
                var state=InitialSessionState.CreateDefault();
                var types=new Dictionary<string,Type>{{"Get",typeof(FeedbackGetBinding)},{"New",typeof(FeedbackNewBinding)},{"Set",typeof(FeedbackSetBinding)},{"Remove",typeof(FeedbackRemoveBinding)}};
                string name="Invoke-Feedback"+verb+"Fixture";state.Commands.Add(new SessionStateCmdletEntry(name,types[verb],null));
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
            try{return JsonConvert.SerializeObject(new{identities=new FeedbackGetBinding().ApplyFilter(pattern),sends=sender.Capture.Sends,tokens=host.TokenRequests});}
            finally{field.SetValue(null,old);HostSingletionInstance.HostContract=oldHost;}
        }
    }
}
