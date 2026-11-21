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
    static class EventsBound {
        public static string Value(PSCmdlet cmdlet,InvocationInfo info,string set,bool? specified=null) {
            var settings=new Dictionary<string,object>();
            foreach(var p in typeof(TeamsEventsPolicy).GetProperties().Where(p=>p.Name!="Identity")) {
                var source=cmdlet.GetType().GetProperty(p.Name);
                if(source!=null)settings.Add(p.Name,source.GetValue(cmdlet));
            }
            return JsonConvert.SerializeObject(new{parameter_set=set,bound=info.BoundParameters,specified,settings});
        }
    }
    [Cmdlet("Invoke","EventsGetFixture",DefaultParameterSetName="Identity")]
    public sealed class EventsGetBinding : GetTeamsEventsPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(EventsBound.Value(this,MyInvocation,ParameterSetName));}
        public string[] ApplyFilter(string pattern){Filter=pattern;return FilterPolicies(new List<TeamsEventsPolicy>{new(){Identity="global"},new(){Identity="Tag:Fixture"},new(){Identity="Global"},new(){Identity=" "},new(){Identity="Global/fixture"}}).Select(p=>p.Identity).ToArray();}
    }
    [Cmdlet("Invoke","EventsNewFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class EventsNewBinding : NewTeamsEventsPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(EventsBound.Value(this,MyInvocation,ParameterSetName,AreParametersSpecified()));}
    }
    [Cmdlet("Invoke","EventsSetFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class EventsSetBinding : SetTeamsEventsPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(EventsBound.Value(this,MyInvocation,ParameterSetName,AreParametersSpecified()));}
    }
    [Cmdlet("Invoke","EventsRemoveFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class EventsRemoveBinding : RemoveTeamsEventsPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(EventsBound.Value(this,MyInvocation,ParameterSetName));}
    }
    [Cmdlet("Invoke","EventsGrantFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class EventsGrantBinding : GrantTeamsEventsPolicyCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){
            bool noScope=false;
            if(ParameterSetName=="Identity") {
                var execute=typeof(BaseTpmGrantCmdlet<TeamsEventsPolicy>).GetMethod("ExecuteGrant",BindingFlags.NonPublic|BindingFlags.Instance);
                var result=(PolicyResponse)execute.Invoke(this,new object[]{"offline-fixture-only","api.interfaces.records.teams.microsoft.com"});
                noScope=result!=null&&!result.IsSuccessStatusCode&&Equals(result.PolicyResult,false);
            }
            WriteObject(JsonConvert.SerializeObject(new{parameter_set=ParameterSetName,bound=MyInvocation.BoundParameters,no_scope_result=noScope}));
        }
    }
    public static class EventsBindingOracle {
        public static string Run(string verb,string arguments){
            if(PowershellUtils.IsSelfHost()||PowershellUtils.IsTorusEnvironment())throw new InvalidOperationException("Public offline profile required");
            var field=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);
            var old=field.GetValue(null);var oldHost=HostSingletionInstance.HostContract;var host=new ProviderFixtureHost();
            using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);field.SetValue(null,client);HostSingletionInstance.HostContract=host;
            string output=null,failure=null;bool hadErrors=false;
            try{
                var state=InitialSessionState.CreateDefault();
                var types=new Dictionary<string,Type>{{"Get",typeof(EventsGetBinding)},{"New",typeof(EventsNewBinding)},{"Set",typeof(EventsSetBinding)},{"Remove",typeof(EventsRemoveBinding)},{"Grant",typeof(EventsGrantBinding)}};
                string name="Invoke-Events"+verb+"Fixture";state.Commands.Add(new SessionStateCmdletEntry(name,types[verb],null));
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
            try{return JsonConvert.SerializeObject(new{identities=new EventsGetBinding().ApplyFilter(pattern),sends=sender.Capture.Sends,tokens=host.TokenRequests});}
            finally{field.SetValue(null,old);HostSingletionInstance.HostContract=oldHost;}
        }
    }
}
