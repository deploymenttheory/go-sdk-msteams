// Pure pinned model setters/serialization. No command or network helper.
using Newtonsoft.Json;
using Microsoft.Teams.Policy.Administration.Cmdlets.Core;
namespace TeamsContractOracle
{
    public static class VoicemailValidationModelOracle
    {
        public static string Run(string uri, bool nullValue)
        {
            var value = new OnlineVoicemailValidationConfiguration { Identity = "Global", AudioFileValidationEnabled = false };
            value.AudioFileValidationUri = nullValue ? null : uri;
            return JsonConvert.SerializeObject(value);
        }
    }
}
