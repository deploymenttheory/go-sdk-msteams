using System;
using System.Collections.Generic;
using System.Net;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using H=Microsoft.Teams.ConfigAPI.CmdletHostContract;
using P=Microsoft.Teams.Policy.Administration.Cmdlets.Providers;
namespace Ext90Offline {
 public sealed class MemoryLogger : H.ILogMessage, P.ILogger {
  public readonly List<string> Events=new List<string>();
  void Capture(string value){lock(Events){Events.Add(value??"");}}
  public void SetLogLevel(bool a,bool b,bool c,bool d){Capture("SetLogLevel");}
  public void ResetLogLevel(){Capture("ResetLogLevel");}
  public void WriteInfo(string s){Capture(s);} public void WriteVerbose(string s){Capture(s);}
  public void WriteWarning(string s){Capture(s);} public void WriteError(string s){Capture(s);}
  public void WriteError(Exception e,string a,string b,string c,string d){Capture(e.GetType().FullName+":"+a);}
  public void WriteInfo(string a,string b,string c,string d,int n){Capture(a+":"+b+":"+c);}
  public void WriteError(string a,string b,string c,string d,int n){Capture(a+":"+b+":"+c);}
  public void WriteMessage(string a,string b,string c,string d,int n){Capture(a+":"+b+":"+c);}
 }
 public sealed class MemoryTelemetry : H.IHostTelemetryClient {
  public readonly List<H.CmdletExecutionInfo> Events=new List<H.CmdletExecutionInfo>();
  public void LogCmdletExecutionInfo(H.CmdletExecutionInfo info){lock(Events){Events.Add(info);}}
 }
 public sealed class FakeHost : H.ICmdletHostContract {
  public H.ILogMessage Logger {get;set;}
  public H.IHostTelemetryClient TelemetryClient {get;set;}
  public int TokenCalls; public int ModuleCalls; public int MaxTokenCalls=8;
  public readonly List<H.TokenGenerationInfo> Requests=new List<H.TokenGenerationInfo>();
  public FakeHost(MemoryLogger logger,MemoryTelemetry telemetry){Logger=logger;TelemetryClient=telemetry;}
  public H.LoginInfo GetAccessToken(H.TokenGenerationInfo info){
   if(Interlocked.Increment(ref TokenCalls)>MaxTokenCalls)throw new InvalidOperationException("Offline token budget exceeded");
   lock(Requests){Requests.Add(info);}
   return new H.LoginInfo{AccessToken="offline-not-a-token",TenantId="00000000-0000-0000-0000-000000000001",AdministeredDomain="offline.example.invalid",TargetTenantId="00000000-0000-0000-0000-000000000001",ConfigApiEnvironment="Prod",ConfigApiEndpoint=new Uri("https://offline.example.invalid/")};
  }
  public void GetModulesInfo(out string configApi,out string teams){Interlocked.Increment(ref ModuleCalls);configApi="0.0.0-offline";teams="8.1.0-offline";}
 }
 public sealed class PlannedReply {
  public string Method; public string AbsoluteUri; public string ExpectedBody;
  public HttpStatusCode Status=HttpStatusCode.OK; public string Json="{}";
 }
 public sealed class TerminalHandler : HttpMessageHandler {
  public readonly Queue<PlannedReply> Replies=new Queue<PlannedReply>();
  public readonly List<string> Captures=new List<string>();
  public int Calls; public int MaxCalls=16;
  protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request,CancellationToken cancellationToken){
   cancellationToken.ThrowIfCancellationRequested();
   if(Interlocked.Increment(ref Calls)>MaxCalls)throw new InvalidOperationException("Offline HTTP budget exceeded");
   if(request.RequestUri==null||request.RequestUri.Scheme!="https"||request.RequestUri.Host!="offline.example.invalid")throw new InvalidOperationException("Unexpected offline request authority");
   string body=request.Content==null?null:request.Content.ReadAsStringAsync().GetAwaiter().GetResult();
   lock(Replies){
    Captures.Add(request.Method.Method+" "+request.RequestUri.AbsoluteUri+"\n"+request.Headers.ToString()+"\n"+(request.Content==null?"":request.Content.Headers.ToString())+"\n"+body);
    if(Replies.Count==0)throw new InvalidOperationException("No offline response scheduled");
    PlannedReply plan=Replies.Dequeue();
    if(request.Method.Method!=plan.Method||request.RequestUri.AbsoluteUri!=plan.AbsoluteUri||(body!=plan.ExpectedBody && !(request.Method==HttpMethod.Get && String.IsNullOrEmpty(body) && String.IsNullOrEmpty(plan.ExpectedBody))))throw new InvalidOperationException("Unexpected offline request vector");
    return Task.FromResult(new HttpResponseMessage(plan.Status){RequestMessage=request,Content=new StringContent(plan.Json,System.Text.Encoding.UTF8,"application/json")});
   }
  }
 }
}
