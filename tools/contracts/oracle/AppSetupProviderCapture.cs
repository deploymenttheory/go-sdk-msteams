// Development-only pinned helper and model observations; no command lifecycle.
using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;
using Newtonsoft.Json;

namespace TeamsContractOracle
{
    public static class AppSetupOracle
    {
        public static readonly string[] Collections = { "AppPresetList", "PinnedAppBarApps", "PinnedMessageBarApps", "AppPresetMeetingList", "AdditionalCustomizationApps", "PinnedCallingBarApps" };
        static Type ElementType(string field) => typeof(TeamsAppSetupPolicy).GetProperty(field).PropertyType.GetGenericArguments()[0];
        static object Entry(string field, string id, long order, string customization) => field switch {
            "AppPresetList" => new AppPreset(id), "PinnedAppBarApps" => new PinnedApp(id,order),
            "PinnedMessageBarApps" => new PinnedMessageBarApp(id,order), "AppPresetMeetingList" => new AppPresetMeeting(id),
            "AdditionalCustomizationApps" => new AdditionalCustomizationApp(id,customization), "PinnedCallingBarApps" => new PinnedCallingBarApp(id),
            _ => throw new ArgumentException("Unknown synthetic collection")
        };

        public static ProviderCapture Run(string operation, string fields, int status, string response, bool force, bool nullPolicyName)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField = typeof(HttpClientHelper).GetField("_client", BindingFlags.Static | BindingFlags.NonPublic);
            var versionField = typeof(PowershellUtils).GetField("moduleVersion", BindingFlags.Static | BindingFlags.NonPublic);
            var oldClient = clientField.GetValue(null); var oldVersion = versionField.GetValue(null);
            var oldHost = HostSingletionInstance.HostContract; var host = new ProviderFixtureHost();
            using var sender = new ProviderSender(status,response); using var client = new HttpClient(sender);
            client.DefaultRequestHeaders.Add("MPACmdlet","true"); clientField.SetValue(null,client); versionField.SetValue(null,"8.1.0"); HostSingletionInstance.HostContract = host;
            try {
                const string identity="scope /Fixture / Child", token="offline-fixture-only", origin="api.interfaces.records.teams.microsoft.com";
                string command=operation switch { "create"=>"New-CsTeamsAppSetupPolicy", "patch"=>"Set-CsTeamsAppSetupPolicy", "delete"=>"Remove-CsTeamsAppSetupPolicy", "user" or "group" or "tenant"=>"Grant-CsTeamsAppSetupPolicy", _=>"Get-CsTeamsAppSetupPolicy" };
                var bound=new Dictionary<string,object>{{"Identity",identity}}; if(force)bound.Add("Force",new SwitchParameter(false));
                var value=new TeamsAppSetupPolicy{Identity="read-model-identity",Description="",AllowSideLoading=false,AllowUserPinning=null};
                if(operation=="create"||operation=="patch") {
                    if(fields=="all"||fields=="scalars"){bound.Add("Description",value.Description);bound.Add("AllowSideLoading",value.AllowSideLoading);bound.Add("AllowUserPinning",value.AllowUserPinning);}
                    if(fields=="all"||fields=="empty-lists")foreach(string field in Collections){
                        var list=(IList)typeof(TeamsAppSetupPolicy).GetProperty(field).GetValue(value);
                        if(fields=="all")list.Add(Entry(field,"fixture-"+field,9007199254740993L,"fixture-customization"));
                        // The helper consumes bound-property membership; the typed
                        // DTO supplies full array contents, not server operators.
                        bound.Add(field,list);
                    }
                }
                var ctor=typeof(InvocationInfo).GetConstructor(BindingFlags.Instance|BindingFlags.NonPublic|BindingFlags.Public,null,new[]{typeof(CommandInfo),typeof(IScriptExtent)},null);
                var invocation=(InvocationInfo)ctor.Invoke(new object[]{new CmdletInfo(command,typeof(GetTeamsAppSetupPolicyCmdlet)),null});foreach(var item in bound)invocation.BoundParameters.Add(item.Key,item.Value);
                string name=nullPolicyName?null:"";
                PolicyResponse result=operation switch{
                    "read"=>HttpRequestHelper<TeamsAppSetupPolicy>.GetAsync(identity,token,origin,invocation).GetAwaiter().GetResult(),
                    "list"=>HttpRequestHelper<TeamsAppSetupPolicy>.GetAllAsync(token,origin,invocation).GetAwaiter().GetResult(),
                    "create"=>HttpRequestHelper<TeamsAppSetupPolicy>.CreateAsync(value,identity,token,origin,invocation,null).GetAwaiter().GetResult(),
                    "patch"=>HttpRequestHelper<TeamsAppSetupPolicy>.SetAsync(value,identity,token,origin,invocation).GetAwaiter().GetResult(),
                    "delete"=>HttpRequestHelper<TeamsAppSetupPolicy>.RemoveAsync(identity,token,origin,invocation).GetAwaiter().GetResult(),
                    "user"=>HttpRequestHelper<TeamsAppSetupPolicy>.GrantToUserAsync("user/fixture",name,token,origin,invocation).GetAwaiter().GetResult(),
                    "group"=>HttpRequestHelper<TeamsAppSetupPolicy>.GrantToGroupAsync("group/fixture",name,token,origin,invocation,-1).GetAwaiter().GetResult(),
                    "tenant"=>HttpRequestHelper<TeamsAppSetupPolicy>.GrantToTenantAsync(name,token,origin,invocation,false).GetAwaiter().GetResult(),
                    _=>throw new ArgumentException("Unknown synthetic operation")
                };
                sender.Capture.Success=result.IsSuccessStatusCode;sender.Capture.DecodedNull=result.PolicyResult==null;sender.Capture.DecodedType=result.PolicyResult?.GetType().FullName;sender.Capture.Warnings=result.Warnings?.ToArray();
            }catch(Exception error){sender.Capture.FailureType=error.GetType().FullName;}
            finally{sender.Capture.TokenRequests=host.TokenRequests;HostSingletionInstance.HostContract=oldHost;versionField.SetValue(null,oldVersion);clientField.SetValue(null,oldClient);}
            return sender.Capture;
        }

        public static string PolicyModel(string input){
            object value=null;string failure=null;
            try{value=input switch{"constructor"=>new TeamsAppSetupPolicy(),"long-description"=>new TeamsAppSetupPolicy{Description=new string('x',2000),AllowSideLoading=false,AllowUserPinning=null},_=>JsonConvert.DeserializeObject<TeamsAppSetupPolicy>(input)};}
            catch(Exception error){failure=error.GetType().FullName;}return JsonConvert.SerializeObject(new{failure,value});
        }
        public static string ElementModel(string field,string input){
            object value=null;string failure=null;
            try{value=input.StartsWith("json:")?JsonConvert.DeserializeObject(input.Substring(5),ElementType(field)):Entry(field,input=="null"?null:input,9007199254740993L,"fixture-customization");}
            catch(Exception error){failure=(error.InnerException??error).GetType().FullName;}return JsonConvert.SerializeObject(new{failure,value});
        }
        public static string Lists(string field,string name){
            Type type=ElementType(field);return (string)typeof(AppSetupOracle).GetMethod("ListCase",BindingFlags.Static|BindingFlags.NonPublic).MakeGenericMethod(type).Invoke(null,new object[]{field,name});
        }
        static string ListCase<T>(string field,string name){
            Func<string,long,string,T> create=(id,order,custom)=>(T)Entry(field,id,order,custom);
            var first=create("a",1,"one");var duplicate=create("a",1,"one");var otherOrder=create("a",2,"one");var otherCustomization=create("a",1,"two");var upper=create("A",1,"one");
            var values=new List<T>{first,duplicate,otherOrder,otherCustomization,upper};string failure=null;
            try{
                PSListModifier<T> edit=name switch{
                    "null"=>null,"empty-operators"=>new PSListModifier<T>(new Hashtable{{"Add",Array.Empty<T>()},{"Remove",Array.Empty<T>()}}),
                    "remove-one"=>new PSListModifier<T>(new Hashtable{{"Remove",new[]{duplicate}}}),
                    "remove-two"=>new PSListModifier<T>(new Hashtable{{"Remove",new[]{duplicate,duplicate}}}),
                    "remove-three"=>new PSListModifier<T>(new Hashtable{{"Remove",new[]{duplicate,duplicate,duplicate}}}),
                    "replace-duplicates"=>new PSListModifier<T>(new object[]{upper,upper,first}),
                    "add-existing"=>new PSListModifier<T>(new Hashtable{{"Add",new[]{duplicate,duplicate}}}),
                    "remove-and-add"=>new PSListModifier<T>(new Hashtable{{"Remove",new[]{duplicate}},{"Add",new[]{duplicate}}}),
                    _=>throw new ArgumentException("Unknown synthetic edit")
                };
                edit.ApplyTo(values,(a,b)=>{
                    string left=(string)typeProperty(field,"Id").GetValue(a),right=(string)typeProperty(field,"Id").GetValue(b);if(left!=right)return false;
                    if(field=="PinnedAppBarApps"||field=="PinnedMessageBarApps")return (long)typeProperty(field,"Order").GetValue(a)==(long)typeProperty(field,"Order").GetValue(b);
                    if(field=="AdditionalCustomizationApps")return (string)typeProperty(field,"AdditionalCustomizationId").GetValue(a)==(string)typeProperty(field,"AdditionalCustomizationId").GetValue(b);
                    return true;
                });
            }catch(Exception error){failure=(error.InnerException??error).GetType().FullName;}
            return JsonConvert.SerializeObject(new{failure,values});
        }
        static PropertyInfo typeProperty(string field,string property)=>ElementType(field).GetProperty(property);
    }
}
