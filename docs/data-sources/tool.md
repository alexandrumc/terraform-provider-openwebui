---
layout: data-source
page_title: "openwebui_tool Data Source"
sidebar_current: docs-openwebui-data-source-tool
description: |-
  Retrieves tool details from Open WebUI.
---

# openwebui_tool (Data Source)

Fetches a tool definition by ID.

## Example Usage

### Minimal

```hcl
data "openwebui_tool" "calculator" {
  id = "calculator"
}
```

### Full

```hcl
data "openwebui_tool" "calculator" {
  id = openwebui_tool.calculator.id
}
```

## Argument Reference

* `id` (Required) – Identifier of the tool to retrieve.

## Attribute Reference

* `id` – Tool identifier.
* `name` – Tool name.
* `content` – Tool content.
* `description` – Tool description.
* `read_groups` / `write_groups` – Access control group names.
* `user_id` – Owner user ID.
* `created_at` / `updated_at` – Timestamps.
* `write_access` – Whether the current user has write access.
