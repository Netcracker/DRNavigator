# sm-dummy

Dummy service to test site-manager

## Helm parameters

    | Parameter                           | Description                                                                    | Default value                                   |
    |-------------------------------------|--------------------------------------------------------------------------------|-------------------------------------------------|
    | name                                | sm-dummy entities name                                                         | "sm-dummy"                                      |
    | PAAS_PLATFORM                       | define PAAS type. It can be "KUBERNETES" or "OPENSHIFT"                        | "KUBERNETES"                                    |
    | env.SMA_TIMEOUT                     | sleep timeout for DR procedure                                                 | 15                                              |
    | env.SMA_INIT_MODE                   | initial mode for service                                                       | "active"                                        |
    | env.SMA_DEBUG                       | debug mode for Flask                                                           | True                                            |
    | env.SMA_HEALTHZ                     | healthz output in dummy mode                                                   | "up"                                            |
    | env.SMA_SEND_MESSAGE                | send messages with /sitemanager request                                        | True                                            |
    | env.SMA_HTTP_AUTH                   | check token for /sitemanager requests                                          | True                                            |
    | env.SMA_SM_NAMESPACE                | site-manager namespace (for authorization checking)                            | "site-manager"                                  |
    | env.SMA_SM_SA                       | site-manager service account name (for authorization checking)                 | "site-manager-sa"                               |
    | env.SMA_CUSTOM_AUDIENCE             | site-manager custom audience for rest API token                                | ""                                              |
    | image                               | sm-dummy image                                                                 |                                                 |
    | ingressName                         | define URL for ingress                                                         |                                                 |
    | serviceAccount.create               | enable/disable Service Account creation                                        | true                                            |
    | serviceAccount.name                 | name of Service Account for sm-dummy                                           | "sm-dummy-sa"                                   |
    | siteManager.instellCR               | enable sitemanager-cr                                                          | True                                            |
    | siteManager.module                  | cr module property                                                             | "stateful"                                      |
    | siteManager.before                  | cr before property                                                             | []                                              |
    | siteManager.after                   | cr after property                                                              | []                                              |
    | siteManager.sequence                | cr sequence property                                                           | ["standby", "active"]                           |
    | siteManager.allowedStandbyStateList | cr allowedStandbyStateList property                                            | ["up"]                                          |
    | siteManager.timeout                 | cr timeout property                                                            | 180                                             |
    | siteManager.useDefaultParams        | add parameters for stateful module in cr (serviceEndpoint and healthzEndpoint) | True                                            |
    | siteManager.additionalParams        | you can customize parameters, adding needed properties inside this param       |                                                 |