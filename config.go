package slogwit

import (
	"fmt"
)

func entryIndexDefaultConfig(indexId string) string {
	return fmt.Sprintf(`{
  "version": "0.8",
  "index_id": %q,
  "doc_mapping": {
    "mode": "lenient",
    "field_mappings": [
      {
        "name": "timestamp",
        "type": "datetime",
        "input_formats": [
          "rfc3339"
        ],
        "output_format": "unix_timestamp_nanos",
        "fast_precision": "nanoseconds",
        "fast": true
      },
      {
        "name": "labels",
        "type": "array<text>"
      },
      {
        "name": "severity",
        "type": "text"
      },
      {
        "name": "payload",
        "type": "text"
      }
    ],
    "timestamp_field": "timestamp"
  },
  "search_settings": {
    "default_search_fields": [
      "severity",
      "payload",
      "labels"
    ]
  },
  "retention": {
    "period": "30 days",
    "schedule": "daily"
  }
}`, indexId)
}
