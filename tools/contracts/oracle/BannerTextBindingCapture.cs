// Inherited compiled parameter binding with synthetic lifecycle overrides.
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
    static class BannerBound {
        public static string Value(InvocationInfo info,string set,string text,string description) => JsonConvert.SerializeObject(new{parameter_set=set,bound=info.BoundParameters,text,description});
    }
    [Cmdlet("Invoke","BannerGetFixture",DefaultParameterSetName="Identity")]
    public sealed class BannerGetBinding : GetTeamsCustomBannerTextCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(BannerBound.Value(MyInvocation,ParameterSetName,null,null));}
        public string[] ApplyFilter(string pattern){Filter=pattern;return FilterPolicies(new List<CustomBannerText>{new(){Identity="global"},new(){Identity="Tag:Fixture"},new(){Identity="Global"},new(){Identity=" "},new(){Identity="Global/fixture"}}).Select(v=>v.Identity).ToArray();}
    }
    [Cmdlet("Invoke","BannerNewFixture",SupportsShouldProcess=true)]
    public sealed class BannerNewBinding : NewTeamsCustomBannerTextCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(BannerBound.Value(MyInvocation,ParameterSetName,Text,Description));}
    }
    [Cmdlet("Invoke","BannerSetFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class BannerSetBinding : SetTeamsCustomBannerTextCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(BannerBound.Value(MyInvocation,ParameterSetName,Text,Description));}
    }
    [Cmdlet("Invoke","BannerRemoveFixture",DefaultParameterSetName="Identity",SupportsShouldProcess=true)]
    public sealed class BannerRemoveBinding : RemoveTeamsCustomBannerTextCmdlet {
        protected override void BeginProcessing(){} protected override void EndProcessing(){}
        protected override void ProcessRecord(){WriteObject(BannerBound.Value(MyInvocation,ParameterSetName,null,null));}
    }
    public static class BannerBindingOracle {
        public static string Run(string verb,string arguments){
            var clientField=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);
            var oldClient=clientField.GetValue(null);var oldHost=HostSingletionInstance.HostContract;
            var host=new ProviderFixtureHost();using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);clientField.SetValue(null,client);HostSingletionInstance.HostContract=host;
            string output=null,failure=null;bool hadErrors=false;
            try{
                var state=InitialSessionState.CreateDefault();
                var types=new Dictionary<string,Type>{{"Get",typeof(BannerGetBinding)},{"New",typeof(BannerNewBinding)},{"Set",typeof(BannerSetBinding)},{"Remove",typeof(BannerRemoveBinding)}};
                string name="Invoke-Banner"+verb+"Fixture";state.Commands.Add(new SessionStateCmdletEntry(name,types[verb],null));
                using var runspace=RunspaceFactory.CreateRunspace(state);runspace.Open();using var ps=PowerShell.Create();ps.Runspace=runspace;
                ps.AddCommand(name);foreach(var p in JsonConvert.DeserializeObject<Dictionary<string,object>>(arguments))ps.AddParameter(p.Key,p.Value);
                var values=ps.Invoke();hadErrors=ps.HadErrors;
                if(hadErrors)failure=ps.Streams.Error.First().Exception.GetType().FullName;
                if(values.Count==1)output=(string)values[0].BaseObject;
                if(!ReferenceEquals(clientField.GetValue(null),client))throw new InvalidOperationException("Constructor replaced memory client");
            }catch(Exception error){failure=(error.InnerException??error).GetType().FullName;}
            finally{clientField.SetValue(null,oldClient);HostSingletionInstance.HostContract=oldHost;}
            return JsonConvert.SerializeObject(new{failure,had_errors=hadErrors,output,sends=sender.Capture.Sends,tokens=host.TokenRequests});
        }
        public static string Filter(string pattern){
            var field=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);var old=field.GetValue(null);
            using var sender=new ProviderSender(200,"{}");using var client=new HttpClient(sender);field.SetValue(null,client);
            try{return JsonConvert.SerializeObject(new{identities=new BannerGetBinding().ApplyFilter(pattern),sends=sender.Capture.Sends});}
            finally{field.SetValue(null,old);}
        }
    }
}
