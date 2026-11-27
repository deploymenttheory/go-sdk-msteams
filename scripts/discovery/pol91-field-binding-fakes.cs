using System.Collections.Generic;
using System.Management.Automation;
namespace Pol91FieldBinding {
public sealed class Observation {
    public Dictionary<string, object> BoundValues = new Dictionary<string, object>();
    public Dictionary<string, object> PropertyValues = new Dictionary<string, object>();
}
[Cmdlet("Test", "Pol91FieldBinding", DefaultParameterSetName="Identity")]
public sealed class Probe : PSCmdlet {
    [Parameter(Mandatory=true, ParameterSetName="Identity")]
    [ValidateNotNullOrEmpty] public string Identity { get; set; }
    [Parameter(ParameterSetName="Identity")] public string RestrictedSenderList { get; set; }
    [Parameter(ParameterSetName="Identity")] public string ContentPin { get; set; }
    [Parameter(ParameterSetName="Identity")] public string ResourceAccountContentAccess { get; set; }
    [Parameter(ParameterSetName="Identity")] public string UseUnifiedDomain { get; set; }
    [Parameter(ParameterSetName="Identity")] public string AudioFileValidationUri { get; set; }
    [Parameter(ParameterSetName="Identity")] public bool? AllowEmailIntoChannel { get; set; }
    [Parameter(ParameterSetName="Identity")] public bool? AudioFileValidationEnabled { get; set; }
    protected override void ProcessRecord() {
        var result = new Observation();
        foreach (var entry in MyInvocation.BoundParameters) {
            result.BoundValues[entry.Key] = entry.Value;
            result.PropertyValues[entry.Key] = GetType().GetProperty(entry.Key).GetValue(this);
        }
        WriteObject(result);
    }
}
}
