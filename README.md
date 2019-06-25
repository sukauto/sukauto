# Sukauto
#### simple service monitor
Check commit
### config example:
{ 

"services": [
    "exmampleSrvNameFirst",
    "exmampleSrvNameTwo",
    "exmampleSrvNameThird"
  ],
  
  "users": {
    "denny": "my_pass_12345",
    "jhon": "abcd98765"
  },
  
  "global": true
  
}


## Check executable environment


* `SERVICE` - service name
* `EVENT` - event name (created, remove, started, stopped, restarted, updated, enabled, disabled) 