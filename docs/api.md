---
layout: default
title: API Explorer
description: Interactive REST API reference for the Nanayam distribution gateway.
---

# API Explorer

An interactive, always-in-sync view of the gateway's REST API, generated from
[`docs/openapi.yaml`]({{ '/openapi.yaml' | relative_url }}). Use **Try it out**
against a gateway running on `localhost:8080` — set a bearer token first (see
below) for anything past `/health`, `/v1/Config`, `/v1/Login`, and
`/v1/Register`.

For the narrative version — authentication flow, status codes, and the
equivalent gRPC service — see the [API Reference](wiki/API-Reference.html)
([தமிழ்](wiki/API-Reference-ta.html)). The raw spec can also be opened
directly: [`openapi.yaml`]({{ '/openapi.yaml' | relative_url }}).

<div id="swagger-ui"></div>

<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: "{{ '/openapi.yaml' | relative_url }}",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout",
      deepLinking: true,
    });
  };
</script>
