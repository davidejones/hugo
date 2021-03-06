---
title: merge
description: "`merge` deep merges two maps and returns the resulting map."
date: 2019-08-08
categories: [functions]
menu:
  docs:
    parent: "functions"
keywords: [dictionary]
signature: ["collections.Merge MAP MAP...", "merge MAP MAP..."]
workson: []
hugoversion: "0.56.0"
relatedfuncs: [dict, append, reflect.IsMap, reflect.IsSlice]
aliases: []
---

Merge creates a copy of the final `MAP` and merges any preceding `MAP` into it in reverse order.
Key handling is case-insensitive.

An example merging two maps.

```go-html-template
{{ $default_params := dict "color" "blue" "width" "50%" "height" "25%" "icon" "star" }}
{{ $user_params := dict "color" "red" "icon" "mail" "extra" (dict "duration" 2) }}
{{ $params := merge $default_params $user_params }}
```

Resulting __$params__:

```
"color": "red"
"extra":
  "duration": 2
"height": "25%"
"icon": "mail"
"width": "50%"
```

{{% note %}}
  Regardless of depth, merging only applies to maps. For slices, use [append]({{< ref "functions/append" >}})
{{% /note %}}
