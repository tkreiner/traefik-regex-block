# traefik-regex-block

## Summary
The traefix-regex-block plugin provides middleware for Traefik to detect URLs that match a list of regex patterns and then block those IP addresses for a configurable amount of time. This project took inspiration from the well known Fail2Ban application. This plugin can improve security of a web site by detecting when threat actors are scanning your site for known exploitable endpoints. If there are certain patterns that are often scanned for that you know will not be on your site, you can proactively identify when those attempts are made and block the source from further scanning of your site.

## Installation
Installation instructions can be found on the [Traefik plugin catalog](https://plugins.traefik.io/plugins/65f7bc4d46079255c9ffd1f0/regex-block).

## Latest Release
The current release is version **v0.0.3**. This plugin is still in it's early development phase. However, it is fully functional and is in bug testing phase. If you encounter any problems, please provide feedback by [opening an issue here](https://github.com/tkreiner/traefik-regex-block/issues).

## Configuration
The following settings can be used to configure your plugin.

### Block Duration - blockDurationMinutes
* Required: No
* Default: 60 minutes

Use this setting to determine how many minutes an IP address will be blocked from your site after each URL attempt that matches a regex pattern.

### Regex Pattern List - regexPatterns
* Required: Yes
* Default: (none)

You provide a list of regex patterns to be used to detect activity you want to block. You can provide any number of patterns to monitor with.

### Whitelist IP Addresses - whitelist
* Required: No
* Default: (none)

If you want to keep from blocking specific IP addresses, you can use the whitelist feature. This accepts a list of IP addresses as either an IP address on in CIDR notation.

### Example
The following configuration will detect any URL traffic that includes `/.env` or `/cgi-bin` in the URL. It will block any further requests from the IP address for 2 hours. It then excludes blocking for any requests from the `127.0.0.1` address, or from a `192.168.0.0/16` network.
```yaml
    my-regex-block:
      plugin:
        traefik-regex-block:
          blockDurationMinutes: 120
          regexPatterns:
            - \/\.env
            - \/cgi-bin
          whitelist:
            - 127.0.0.1
            - 192.168.0.0/16
```
