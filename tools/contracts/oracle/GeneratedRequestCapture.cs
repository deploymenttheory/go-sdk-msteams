// Development-only in-memory adapter for the pinned module's public interfaces.
// No HttpClient, credential provider, module pipeline or management cmdlet is used.
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated;
using Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Runtime;

namespace TeamsContractOracle
{
    public sealed class Capture
    {
        public string Method { get; set; }
        public string GeneratedUrl { get; set; }
        public string ContentType { get; set; }
        public string RequestBody { get; set; }
        public Dictionary<string, string[]> RequestHeaders { get; set; }
        public string Handler { get; set; }
        public string DecodedType { get; set; }
        public bool? DecodedNull { get; set; }
        public string DecoderFailure { get; set; }
        public int Sends { get; set; }
    }

    public sealed class Sender : ISendAsync
    {
        readonly int status;
        readonly string body;
        public readonly Capture Result = new Capture();
        public Sender(int status, string body) { this.status = status; this.body = body; }
        public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, IEventListener listener)
        {
            if (request.Headers.Authorization != null || request.Headers.Contains("Proxy-Authorization"))
                throw new InvalidOperationException("Unexpected credential header in offline oracle");
            Result.Sends++;
            Result.Method = request.Method.Method;
            Result.GeneratedUrl = request.RequestUri.AbsoluteUri;
            Result.ContentType = request.Content?.Headers.ContentType?.ToString();
            Result.RequestBody = request.Content == null ? null : await request.Content.ReadAsStringAsync();
            Result.RequestHeaders = request.Headers.ToDictionary(pair => pair.Key, pair => pair.Value.ToArray(), StringComparer.OrdinalIgnoreCase);
            return new HttpResponseMessage((HttpStatusCode)status) { Content = new StringContent(body ?? "", System.Text.Encoding.UTF8, "application/json") };
        }
    }

    public sealed class Listener : IEventListener
    {
        public CancellationToken Token { get; }
        public Action Cancel => () => { };
        public Listener(bool canceled) { Token = new CancellationToken(canceled); }
        public Task Signal(string id, CancellationToken token, Func<EventData> createMessage) => Task.CompletedTask;
    }

    public sealed class Callback
    {
        readonly Capture result;
        readonly string name;
        public Callback(Capture result, string name) { this.result = result; this.name = name; }
        public Task WithoutBody(HttpResponseMessage response)
        {
            result.Handler = name;
            return Task.CompletedTask;
        }
        public async Task WithBody<T>(HttpResponseMessage response, Task<T> body)
        {
            result.Handler = name;
            try { T model = await body; result.DecodedType = typeof(T).Name; result.DecodedNull = (object)model == null; }
            catch { result.DecoderFailure = "decode-rejected"; }
        }
    }

    public static class Oracle
    {
        public static Capture Run(string operation, object[] inputs, int status, string body, bool canceled)
        {
            MethodInfo method = typeof(ConfigApiBasedCmdlets).GetMethods().Single(m => m.Name == operation);
            var sender = new Sender(status, body);
            var listener = new Listener(canceled);
            var parameters = method.GetParameters();
            var values = new object[parameters.Length];
            int next = 0;
            for (int i = 0; i < parameters.Length; i++)
            {
                Type type = parameters[i].ParameterType;
                if (type == typeof(IEventListener)) values[i] = listener;
                else if (type == typeof(ISendAsync)) values[i] = sender;
                else if (typeof(Delegate).IsAssignableFrom(type))
                {
                    var callback = new Callback(sender.Result, parameters[i].Name);
                    var generic = type.GetGenericArguments();
                    MethodInfo handle = generic.Length == 2
                        ? typeof(Callback).GetMethod("WithoutBody")
                        : typeof(Callback).GetMethod("WithBody").MakeGenericMethod(generic[1].GetGenericArguments()[0]);
                    values[i] = handle.CreateDelegate(type, callback);
                }
                else values[i] = inputs[next++];
            }
            if (next != inputs.Length) throw new ArgumentException("Oracle input count mismatch");
            ((Task)method.Invoke(new ConfigApiBasedCmdlets(), values)).GetAwaiter().GetResult();
            return sender.Result;
        }
    }
}
