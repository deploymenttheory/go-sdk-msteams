// Owned development adapter: actual constructors/scope preparation/direct helper.
// Constructor transport creation is blocked by an already injected memory client.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Management.Automation;
using System.Management.Automation.Language;
using System.Net.Http;
using System.Reflection;
using Newtonsoft.Json;
using Microsoft.Teams.ConfigAPI.CmdletHostContract;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
using Microsoft.Teams.Policy.Administration.Cmdlets.Providers;

namespace TeamsContractOracle
{
    public static class BannerTextOracle
    {
        static Type CommandType(string operation) => operation switch {
            "create" => typeof(NewTeamsCustomBannerTextCmdlet), "patch" => typeof(SetTeamsCustomBannerTextCmdlet),
            "delete" => typeof(RemoveTeamsCustomBannerTextCmdlet), _ => typeof(GetTeamsCustomBannerTextCmdlet)
        };
        static object Prepare(string operation, string selector, string identity, string relativeID, out string normalized,
            out string selectedID, out Dictionary<string,Dictionary<string,object>> metadata, out bool defaultID)
        {
            var type = CommandType(operation);
            var command = Activator.CreateInstance(type);
            defaultID = operation == "create" && Guid.TryParse((string)type.GetProperty("Id").GetValue(command), out _);
            if (selector == "identity") type.GetProperty("Identity").SetValue(command,identity);
            else if (selector == "relative") type.GetProperty("Id").SetValue(command,relativeID);
            else if (selector != "default") throw new ArgumentException("Unknown selector");
            type.GetMethod("PrepareForScopeCheck",BindingFlags.Instance|BindingFlags.NonPublic).Invoke(command,null);
            normalized = (string)type.GetProperty("Identity").GetValue(command);
            selectedID = (string)type.GetProperty("Id").GetValue(command);
            metadata = (Dictionary<string,Dictionary<string,object>>)type.GetProperty("CmdletMetadata").GetValue(command);
            return command;
        }
        public static object Run(string operation, string selector, string identity, string relativeID, string input,
            string readIdentity, string readID, int status, string response, int forceMode)
        {
            if (PowershellUtils.IsSelfHost() || PowershellUtils.IsTorusEnvironment()) throw new InvalidOperationException("Public offline profile required");
            var clientField=typeof(HttpClientHelper).GetField("_client",BindingFlags.Static|BindingFlags.NonPublic);
            var versionField=typeof(PowershellUtils).GetField("moduleVersion",BindingFlags.Static|BindingFlags.NonPublic);
            var oldClient=clientField.GetValue(null);var oldVersion=versionField.GetValue(null);var oldHost=HostSingletionInstance.HostContract;
            var host=new ProviderFixtureHost();using var sender=new ProviderSender(status,response);using var client=new HttpClient(sender);
            client.DefaultRequestHeaders.Add("MPACmdlet","true");clientField.SetValue(null,client);versionField.SetValue(null,"8.1.0");HostSingletionInstance.HostContract=host;
            string normalized=null,selectedID=null;Dictionary<string,Dictionary<string,object>> metadata=null;bool defaultID=false;
            try {
                var commandObject=Prepare(operation,selector,identity,relativeID,out normalized,out selectedID,out metadata,out defaultID);
                if (!ReferenceEquals(clientField.GetValue(null),client)) throw new InvalidOperationException("Constructor replaced memory transport");
                string command=operation switch {"create"=>"New-CsTeamsCustomBannerText","patch"=>"Set-CsTeamsCustomBannerText","delete"=>"Remove-CsTeamsCustomBannerText",_=>"Get-CsTeamsCustomBannerText"};
                var bound=new Dictionary<string,object>();
                if(selector=="identity")bound.Add("Identity",identity);
                if(selector=="relative")bound.Add("Id",relativeID);
                if(forceMode>=0)bound.Add("Force",new SwitchParameter(forceMode==1));
                var settings=JsonConvert.DeserializeObject<Dictionary<string,string>>(input);
                var value=new CustomBannerText();
                if(operation=="create" || operation=="patch") {
                    foreach(var field in settings){
                        if(field.Key!="Text" && field.Key!="Description")throw new ArgumentException("Unknown setting");
                        bound.Add(field.Key,field.Value);
                        typeof(CustomBannerText).GetProperty(field.Key).SetValue(value,field.Value);
                        CommandType(operation).GetProperty(field.Key).SetValue(commandObject,field.Value);
                    }
                    if(operation=="create") {
                        value=(CustomBannerText)CommandType(operation).GetMethod("CreateNew",BindingFlags.Instance|BindingFlags.NonPublic).Invoke(commandObject,null);
                        if(settings.TryGetValue("Description",out var description))value.Description=description;
                    }
                }
                value.Identity=readIdentity;
                if(operation=="patch" && !string.IsNullOrEmpty(readID))value.Id=readID;
                var ctor=typeof(InvocationInfo).GetConstructor(BindingFlags.Instance|BindingFlags.NonPublic|BindingFlags.Public,null,new[]{typeof(CommandInfo),typeof(IScriptExtent)},null);
                var invocation=(InvocationInfo)ctor.Invoke(new object[]{new CmdletInfo(command,CommandType(operation)),null});foreach(var item in bound)invocation.BoundParameters.Add(item.Key,item.Value);
                const string token="offline-fixture-only",origin="api.interfaces.records.teams.microsoft.com";
                PolicyResponse result=operation switch {
                    "read"=>HttpRequestHelper<CustomBannerText>.GetAsync(normalized,token,origin,invocation).GetAwaiter().GetResult(),
                    "list"=>HttpRequestHelper<CustomBannerText>.GetAllAsync(token,origin,invocation).GetAwaiter().GetResult(),
                    "create"=>HttpRequestHelper<CustomBannerText>.CreateAsync(value,normalized,token,origin,invocation,metadata).GetAwaiter().GetResult(),
                    "patch"=>HttpRequestHelper<CustomBannerText>.SetAsync(value,normalized,token,origin,invocation).GetAwaiter().GetResult(),
                    "delete"=>HttpRequestHelper<CustomBannerText>.RemoveAsync(normalized,token,origin,invocation).GetAwaiter().GetResult(),
                    _=>throw new ArgumentException("Unknown operation")
                };
                sender.Capture.Success=result.IsSuccessStatusCode;sender.Capture.DecodedNull=result.PolicyResult==null;sender.Capture.DecodedType=result.PolicyResult?.GetType().FullName;sender.Capture.Warnings=result.Warnings?.ToArray();
            }catch(Exception error){sender.Capture.FailureType=(error.InnerException??error).GetType().FullName;}
            finally{sender.Capture.TokenRequests=host.TokenRequests;clientField.SetValue(null,oldClient);versionField.SetValue(null,oldVersion);HostSingletionInstance.HostContract=oldHost;}
            return new {capture=sender.Capture,normalized_identity=normalized,relative_id=selectedID,metadata,constructor_has_guid=defaultID};
        }
        public static string Model(string input) {
            object value=null;string failure=null;
            try{value=input switch {"constructor"=>new CustomBannerText(),"text-constructor-null"=>new CustomBannerText(null),"text-constructor-empty"=>new CustomBannerText(""),_=>JsonConvert.DeserializeObject<CustomBannerText>(input)};}
            catch(Exception error){failure=(error.InnerException??error).GetType().FullName;}
            return JsonConvert.SerializeObject(new{failure,value});
        }
    }
}
