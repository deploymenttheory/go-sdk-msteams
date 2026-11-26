using System;using System.Collections.Generic;using System.Management.Automation;
using M=Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models;
namespace Ext90InertRealReplay {
public sealed class Observation {public string BodyJson;public Dictionary<string,object> BoundValues=new Dictionary<string,object>();public Dictionary<string,string> BoundTypes=new Dictionary<string,string>();public Dictionary<string,string> BoundText=new Dictionary<string,string>();}
[Cmdlet("Test","Ext90CustomAppBinding",DefaultParameterSetName="SetExpanded")]
public sealed class BindingCmdlet:PSCmdlet {
private M.UpdateCustomAppSettingRequest body=new M.UpdateCustomAppSettingRequest();
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsAppsEnabled {get{return new SwitchParameter(body.IsAppsEnabled.GetValueOrDefault());} set{body.IsAppsEnabled=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsAppsPurchaseEnabled {get{return new SwitchParameter(body.IsAppsPurchaseEnabled.GetValueOrDefault());} set{body.IsAppsPurchaseEnabled=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsExternalAppsEnabledByDefault {get{return new SwitchParameter(body.IsExternalAppsEnabledByDefault.GetValueOrDefault());} set{body.IsExternalAppsEnabledByDefault=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsLicenseBasedPinnedAppsEnabled {get{return new SwitchParameter(body.IsLicenseBasedPinnedAppsEnabled.GetValueOrDefault());} set{body.IsLicenseBasedPinnedAppsEnabled=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsSideloadedAppsInteractionEnabled {get{return new SwitchParameter(body.IsSideloadedAppsInteractionEnabled.GetValueOrDefault());} set{body.IsSideloadedAppsInteractionEnabled=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public SwitchParameter IsTenantWideAutoInstallEnabled {get{return new SwitchParameter(body.IsTenantWideAutoInstallEnabled.GetValueOrDefault());} set{body.IsTenantWideAutoInstallEnabled=new Nullable<bool>(value.IsPresent);}}
[Parameter(ParameterSetName="SetExpanded")] public string LobTextColor {get{return body.LobTextColor;} set{body.LobTextColor=value;}}
[Parameter(ParameterSetName="SetExpanded")] public string LobBackground {get{return body.LobBackground;} set{body.LobBackground=value;}}
[Parameter(ParameterSetName="SetExpanded")] public string LobLogo {get{return body.LobLogo;} set{body.LobLogo=value;}}
[Parameter(ParameterSetName="SetExpanded")] public string LobLogomark {get{return body.LobLogomark;} set{body.LobLogomark=value;}}
[Parameter(ParameterSetName="SetExpanded")] public M.IAppAccessRequestConfig AppAccessRequestConfig {get{return body.AppAccessRequestConfig;}set{body.AppAccessRequestConfig=value;}}
[Parameter(ParameterSetName="SetExpanded"),AllowEmptyCollection] public M.IApplicationObject[] AppSettingsList {get{return body.AppSettingsList;}set{body.AppSettingsList=value;}}
protected override void ProcessRecord(){var result=new Observation{BodyJson=body.ToJsonString()};foreach(var kv in MyInvocation.BoundParameters){result.BoundValues[kv.Key]=kv.Value;result.BoundTypes[kv.Key]=kv.Value==null?"null":kv.Value.GetType().FullName;result.BoundText[kv.Key]=kv.Value==null?null:kv.Value.ToString();}WriteObject(result);}
}}
