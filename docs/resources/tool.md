---
layout: resource
page_title: "openwebui_tool Resource"
sidebar_current: docs-openwebui-resource-tool
description: |-
  Manages Open WebUI tools.
---

# openwebui_tool (Resource)

Creates and manages tool definitions in Open WebUI.

## Example Usage

### Minimal

```hcl
resource "openwebui_tool" "calculator" {
  id      = "calculator"
  name    = "Calculator"
  content = file("./tools/calculator.py")
}
```

### Full

```hcl
resource "openwebui_tool" "calculator" {
  id      = "calculator"
  name    = "Calculator"
  content = file("./tools/calculator.py")

  description   = "Internal calculator tool"
  manifest_json = jsonencode({
    version = "1.0.0"
  })

  read_groups  = ["Support"]
  write_groups = ["Support"]
}
```

## Argument Reference

* `id` (Required) – Identifier for the tool.
* `name` (Required) – Display name for the tool.
* `content` (Required) – Tool source content.
* `description` (Optional) – Human-readable description.
* `manifest_json` (Optional) – JSON manifest for the tool.
* `read_groups` (Optional) – Group names or IDs granted read access.
* `write_groups` (Optional) – Group names or IDs granted write access.

## Attribute Reference

* `id` – Tool identifier (same as input).
* `specs_json` – JSON specification returned by Open WebUI.
* `user_id` – Identifier of the user who owns the tool.
* `created_at` – Unix timestamp of tool creation.
* `updated_at` – Unix timestamp of last update.
* `write_access` – Whether the current user has write access.

## Import

Tools can be imported by their tool ID:

```bash
terraform import openwebui_tool.calculator calculator
```
