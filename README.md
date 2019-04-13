# SOCKSTun – Tunnel TCP sockets through a SOCKS proxy

[![Build Status](https://travis-ci.org/leonklingele/sockstun.svg?branch=master)](https://travis-ci.org/leonklingele/sockstun)

`sockstun` allows to proxy TCP packets from one socket to another through a SOCKS proxy.

## Why this is useful

Some applications such as Apple's Mail app don't support running with `torsocks`.
If you prefer to connect to your mail server through Tor for privacy reasons, `sockstun` will help you.

## Installation

```sh
go get -u github.com/leonklingele/sockstun/...
sockstun -help
```

## Setup

### Overview

In this setup guide we want to proxy TCP traffic reaching the following local ports through a Tor SOCKS proxy running at `localhost:9125`:

- `localhost:1587` to `mail.leonklingele.de:587` (SMTP submission)
- `localhost:1993` to `mail.leonklingele.de:993` (IMAP)

### Setup guide

First, ensure the domain you want to connect to resolves to localhost, in our case:

```sh
$ echo "127.0.0.1 mail.leonklingele.de" | sudo tee -a /etc/hosts
```

Then, edit `sockstun`'s config file:

```sh
$ cat ~/.sockstun/config.toml
# SOCKS proxy URI
socks_uri  = "socks5://localhost:9125"
# Read and write timeout, "0s" for no timeout
rw_timeout = "0s"

# Rule set
[rules]
[rules.mail-leonklingele-imap]
local  = "localhost:1993"
remote = "mail.leonklingele.de:993"
[rules.mail-leonklingele-submission]
local  = "localhost:1587"
remote = "mail.leonklingele.de:587"
```

__Note__: If using a Tor SOCKS proxy, remember to [not mix modes of anonymity](https://trac.torproject.org/projects/tor/wiki/doc/TorifyHOWTO#Remember:Modesofanonymitydonotmix)!
See section [Setting up multiple Tor sessions](#setting-up-multiple-tor-sessions) on how to set up another Tor session.

Now simply start `sockstun`:

```sh
$ sockstun
enabling proxy rule mail-leonklingele-submission (localhost:1587->mail.leonklingele.de:587)
enabling proxy rule mail-leonklingele-imap (localhost:1993->mail.leonklingele.de:993)
```

To test the setup:

```sh
$ openssl s_client -connect mail.leonklingele.de:1993
[..]
* OK [CAPABILITY IMAP4rev1 LITERAL+ SASL-IR LOGIN-REFERRALS ID ENABLE IDLE AUTH=PLAIN] Dovecot ready.
fucksy wucksie!!
fucksy BAD Error in IMAP command received by server.
```

Requests to `mail.leonklingele.de:1993` are now being proxied through Tor.

In order for Apple Mail to actually use the new setup, simply open Mail's preferences, and edit your account as follows:

Use Port `1993` instead of `993`:

![mail-settings-imap](https://www.leonklingele.de/sockstun/mail-settings-imap.png?20190212)

Use Port `1587` instead of `587`:

![mail-settings-submission](https://www.leonklingele.de/sockstun/mail-settings-submission.png?20190212)

#### Using Tor Hidden Services (.onion domains)

`sockstun` works with Tor Hidden Services, too!

```sh
$ cat ~/.sockstun/config.toml
# SOCKS proxy URI
socks_uri  = "socks5://localhost:9125"
# Read and write timeout, "0s" for no timeout
rw_timeout = "0s"

# Rule set
[rules]
[rules.mail-leonklingele-imap]
local  = "localhost:1993"
remote = "gdgd33bknbocpyno5fraft356hh6hc7xaejchjjh4cxnbhxzduq2fiqd.onion:993"
[rules.mail-leonklingele-submission]
local  = "localhost:1587"
remote = "gdgd33bknbocpyno5fraft356hh6hc7xaejchjjh4cxnbhxzduq2fiqd.onion:587"
```

### Setting up multiple Tor sessions

It is advised to not use the same Tor instance for traffic of two different anonymity classes. Please read https://trac.torproject.org/projects/tor/wiki/doc/TorifyHOWTO#Remember:Modesofanonymitydonotmix before continuing.

Follow these steps to launch another instance of Tor with a SOCKS proxy listening at `localhost:9125`:

```sh
$ mkdir -p /usr/local/etc/tor/mail
$ cat <<EOF > /usr/local/etc/tor/mail/torrc
	DataDirectory /usr/local/etc/tor/mail
	SocksPort 127.0.0.1:9125
EOF
$ tor -f /usr/local/etc/tor/mail/torrc
```
