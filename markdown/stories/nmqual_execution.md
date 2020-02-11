## Babylon should allow for NonMem Execution via NMQual
**Product Risk**: low

### Summary
Since NMQual has been used for an extended period of time, the option should be made available to execute nonmem the 
same way that autolog.pl did. By using the `--nmqual=true` flag on the run command, Babylon will specify autolog.pl 
syntax in its executable script, rather than the typical calls directly to nmfe.

#### Tests

##### Automated Tests

Test | Test Name | Count
-----|-----------|-------
nmqual_test.go| TestNMQUALExecutionSucceeds |1 