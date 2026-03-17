# Dink filter

Dink filter acts as a gatekeeper proxy application to which [Dink plugin](https://github.com/pajlads/DinkPlugin)'s requests can be pointed and then based on configuration-defined filter rules, decide which requests should go to which destination.

Currently supported event types are: `Loot`, `Kill Count`.

# Configuring Dink plugin

In Dink's settings, navigate to `Webhook Overrides` section and inside `Loot` and/or `Kill Count` webhook override fields, specify the Dink filter's endpoint only - all final destinations should be specified inside the configuration file.

Example:

![Webhook override example in Dink's settings](https://cdn.zneix.eu/dax75Lz.png)

# Configuring Dink filter application

To run the application, `config.json` configuration file has to be present. An example file `config-dist.json` is provided.

Option            | Description
----------------- | -------------------------------------------------------------------
`routePrefix`     | If requests to the API are coming from e.g. nginx, set this to the path that preceedes requests, e.g. if the API is available under `https://example.com/dink-filter` set to `/dink-filter`
`bindAddress`     | IP and port on which the application will listen for incoming HTTP requests
`password`        | A password that needs to be provided as the `password` URL parameter for incoming requests
`globalFilter`    | Global filter rules that will be used as fallback if destination has no configuration specific to an incoming Dink's request
`destinations`    | Object defining list of destinations, to which requests will be forwarded, as well as their specific rules

### Destination rule configuration

For `globalFilter`, as well as each value of `destinations`, following rules can be used:  
Rules beginning with `enable...` have no effect in `globalFilter` section.

Option                     | Description
------------------------   | -------------------------------------------------------------------
`enableLoot`               | Whether Dink Requests of type `Loot` should be forwarded to this destination
`enableKillCount`          | Whether Dink Requests of type `Kill Count` should be forwarded to this destination
`enableKillCountRegular`   | Whether Dink Requests of type `Kill Count` for non-Personal Best kills should be forwarded to this destination
`enableKillCountPBs`       | Whether Dink Requests of type `Kill Count` for Personal Best kills should be forwarded to this destination
`lootThreshold`            | Required minimum value of all looted items in the `Loot` request to forward the request
`defaultKillCountInterval` | Kill Count interval for which requests should be forwarded if there's no boss-specific interval set (see below)
`killCountIntervals`       | Object with boss-specific intervals, intervals should be specified with key being the Boss' full name and value the interval itself

### Rule priorities

For global & destination-specific rules, the following order is used to deremine loot tresholds:

For `Loot` requests:
1. Check for url-specific loot treshold.
2. As fallback, use global default value. One is required to be present.

For `Kill Count` requests:

1. Check for url-specific boss-specific intervals.
2. Check for global boss-specific intervals.
3. With no boss-specific intervals, check for url-specific default interval.
4. As fallback, use global default interval. One is required to be present.

# Seting up Dink filter application

Dink filter needs to be self-hosted as a server application, [go 1.26.1+](https://go.dev/dl/) is required to build the project and it is recommended to put Dink filter behind a reverse proxy such as nginx or caddy.

The `config.json` configuration file has to be present and properly set up in order for application to start, see 'Configuring Dink filter application' section above.

To build the project, use `make`.

An example systemd unit `dink-filter.service` can be used to run the application, install it with `sudo cp dink-filter.service /etc/systemd/system/` and when ready, start it with `sudo systemctl enable --now dink-filter.service`.


For setting up nginx, the following can be put inside the `server` block in your domain's configuration file. For how to set up nginx in the first place, see [nginx guide](https://nginx.org/en/docs/beginners_guide.html) or ask me a question and I'll try my best to answer.
```conf
location /dink-filter {
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_pass http://localhost:9405;
}
```

