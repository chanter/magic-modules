package datalineage

import (
	"reflect"
	"testing"
)

func TestUnitDataLineageOpenLineageJob_buildRunEventColumnLineageFacet(t *testing.T) {
	t.Parallel()

	in := map[string]interface{}{
		"namespace": "testNamespace",
		"name":      "testJob",
		"input": []interface{}{
			map[string]interface{}{
				"namespace": "bigquery",
				"name":      "my-project.raw.source_table",
			},
		},
		"output": []interface{}{
			map[string]interface{}{
				"namespace": "bigquery",
				"name":      "my-project.curated.target_table",
				"catalog": map[string]interface{}{
					"framework": "iceberg",
					"type":      "bigquery",
					"name":      "example-catalog",
				},
				"symlink": []interface{}{
					map[string]interface{}{
						"namespace": "gs://example-bucket/",
						"name":      "warehouse/curated/target_table",
						"type":      "TABLE",
					},
				},
				"column_lineage": map[string]interface{}{
					"dataset_input": []interface{}{
						map[string]interface{}{
							"namespace": "bigquery",
							"name":      "my-project.raw.source_table",
							"field":     "a",
							"transformation": []interface{}{
								map[string]interface{}{
									"type":    "INDIRECT",
									"subtype": "FILTER",
								},
							},
						},
					},
					"field": []interface{}{
						map[string]interface{}{
							"name": "ident",
							"input": []interface{}{
								map[string]interface{}{
									"namespace": "bigquery",
									"name":      "my-project.raw.source_table",
									"field":     "a",
									"transformation": []interface{}{
										map[string]interface{}{
											"type":    "DIRECT",
											"subtype": "IDENTITY",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	event := buildRunEvent(in)

	outputs, ok := event["outputs"].([]map[string]interface{})
	if !ok || len(outputs) != 1 {
		t.Fatalf("expected 1 output, got %#v", event["outputs"])
	}

	facets, ok := outputs[0]["facets"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected output facets map, got %#v", outputs[0]["facets"])
	}

	if _, exists := facets["column_lineage"]; exists {
		t.Fatalf("unexpected snake_case facet key \"column_lineage\" in %#v", facets)
	}

	cllFacet, ok := facets["columnLineage"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected camelCase facet key \"columnLineage\", got %#v", facets)
	}

	expectedCll := map[string]interface{}{
		"_producer":  openLineageProducer,
		"_schemaURL": openLineageColumnLineageURL,
		"dataset": []map[string]interface{}{
			{
				"namespace": "bigquery",
				"name":      "my-project.raw.source_table",
				"field":     "a",
				"transformations": []map[string]interface{}{
					{
						"type":    "INDIRECT",
						"subtype": "FILTER",
					},
				},
			},
		},
		"fields": map[string]interface{}{
			"ident": map[string]interface{}{
				"inputFields": []map[string]interface{}{
					{
						"namespace": "bigquery",
						"name":      "my-project.raw.source_table",
						"field":     "a",
						"transformations": []map[string]interface{}{
							{
								"type":    "DIRECT",
								"subtype": "IDENTITY",
							},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(cllFacet, expectedCll) {
		t.Errorf("columnLineage facet mismatch:\ngot:  %#v\nwant: %#v", cllFacet, expectedCll)
	}

	if _, ok := facets["catalog"].(map[string]interface{}); !ok {
		t.Errorf("expected catalog facet in output facets, got %#v", facets)
	}
	if _, ok := facets["symlinks"].(map[string]interface{}); !ok {
		t.Errorf("expected symlinks facet in output facets, got %#v", facets)
	}
}
