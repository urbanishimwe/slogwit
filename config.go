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

//  Below codes will be needed in not so distant future

// type searchQueries struct {
// 	query          string
// 	searchFields   []string
// 	startTimestamp time.Time
// 	endTimestamp   time.Time
// 	maxHits        uint64
// 	startOffset    uint64
// 	sortByField    string
// }

// func searchQueriesToString(urlQueries searchQueries) string {
// 	queryBuilder := url.Values{}
// 	if urlQueries.query == "" {
// 		// query named "query" is mandatory
// 		queryBuilder.Add("query", "*")
// 	} else {
// 		queryBuilder.Add("query", urlQueries.query)
// 	}

// 	if !urlQueries.startTimestamp.IsZero() {
// 		queryBuilder.Add("start_timestamp", strconv.Itoa(int(urlQueries.startTimestamp.Unix())))
// 	}
// 	if !urlQueries.endTimestamp.IsZero() {
// 		queryBuilder.Add("end_timestamp", strconv.Itoa(int(urlQueries.startTimestamp.Unix())))
// 	}

// 	if urlQueries.maxHits > 0 {
// 		queryBuilder.Add("max_hits", strconv.Itoa(int(urlQueries.maxHits)))
// 	}

// 	if urlQueries.startOffset > 0 {
// 		queryBuilder.Add("start_offset", strconv.Itoa(int(urlQueries.startOffset)))
// 	}

// 	if len(urlQueries.searchFields) > 0 {
// 		queryBuilder.Add("search_field", strings.Join(urlQueries.searchFields, ","))
// 	}

// 	if urlQueries.sortByField != "" {
// 		queryBuilder.Add("sort_by", urlQueries.sortByField)
// 	}

// 	return queryBuilder.Encode()
// }
