# Assignment migrations

`Client.TeamsAdminAPI.AssignmentMigrations` uses the compiled ConfigAPI GPA protocol.

| SDK method | MicrosoftTeams 8.1.0 command |
|---|---|
| List / GetByID | Get-DirectToGroupAssignmentsMigrationStatus |
| RecommendationsByPolicyType | Get-GroupAssignmentRecommendationsPerPolicyType |
| RecommendationsByPolicyName | Get-GroupAssignmentRecommendationsPerPolicyName |
| GetConflicts | Get-GroupPolicyAssignmentConflict |
| Create | Invoke-StartDirectToGroupAssignmentMigration |
| ClearDirectAssignments | Invoke-ClearDirectToGroupAssignmentMigration |

Each command has the observed `__AllParameterSets` set. There is no evidenced job
Update/Delete endpoint. Cleanup changes direct assignments; it does not delete a
group or job. Acceptance is separate from completion/effective policy propagation.
Neither submission retries, waits, terminal-state inference nor rollback are implicit.
After an ambiguous submission, retain the response and investigate through reads.

Submission requires Authority/PolicyType/PolicyName strings and a canonical GUID
GroupID, normalized to lowercase like bound System.Guid. AllowNonEmptyGroup always
appears as a True/False query string; body keys are authority/policytype/policyname/groupid.
Nil recommendation GroupThreshold uses the string `500`; a supplied pointer preserves
the exact string, including empty. It is not a numeric limit or paging cursor.

Dynamic path segments and query values are encoded separately. Selected-cloud
Teams origin/token, MPACmdlet, command name and fresh correlation metadata apply.
These direct calls do not use generated forest discovery or Graph group lookup;
shared scope audience/destination guards still apply.

Jobs retain open string/Int32/null states, original timestamp text and raw future
fields. Missing CLR defaults are not manufactured. Recommendations/conflicts retain
nested snapshots. Native HTTP/decode errors retain the Resty response. Empty/malformed
successful bodies are decode errors; explicit JSON null returns nil. Actual 403/502
status is preserved instead of the cmdlet's token-invalid/group-in-use output string.

Source requests the selected ConfigAPI resource's delegated user_impersonation scope.
Application eligibility, feature activation, RBAC and cloud deployment need separate
qualification; synthetic mode tests only establish routing. Both current dedicated-SPN
read attempts returned HTTP 403. See [evidence](../../../docs/implementation/policies/assignment-migration-sdk-progress.md)
and the [lab plan](../../../docs/implementation/policies/assignment-migration-lab-acceptance.md).
