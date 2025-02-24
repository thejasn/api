package markers

import (
	"encoding/json"
	"os"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var RequiredFeatureSets = sets.NewString()

func init() {
	featureSet := os.Getenv("OPENSHIFT_REQUIRED_FEATURESET")
	if len(featureSet) == 0 {
		return
	}

	for _, curr := range strings.Split(featureSet, ",") {
		RequiredFeatureSets.Insert(curr)
	}
}

const OpenShiftFeatureSetMarkerName = "openshift:enable:FeatureSets"
const OpenShiftFeatureSetAwareEnumMarkerName = "openshift:validation:FeatureSetAwareEnum"
const OpenShiftFeatureSetAwareXValidationMarkerName = "openshift:validation:FeatureSetAwareXValidation"

func init() {
	ValidationMarkers = append(ValidationMarkers,
		must(markers.MakeDefinition(OpenShiftFeatureSetAwareEnumMarkerName, markers.DescribesField, FeatureSetEnum{})).
			WithHelp(markers.SimpleHelp("OpenShift", "specifies the FeatureSet that is required to generate this field.")),
	)
	FieldOnlyMarkers = append(FieldOnlyMarkers,
		must(markers.MakeDefinition(OpenShiftFeatureSetMarkerName, markers.DescribesField, []string{})).
			WithHelp(markers.SimpleHelp("OpenShift", "specifies the FeatureSet that is required to generate this field.")),
	)
	ValidationMarkers = append(ValidationMarkers,
		must(markers.MakeDefinition(OpenShiftFeatureSetAwareXValidationMarkerName, markers.DescribesType, FeatureSetXValidation{})).
			WithHelp(markers.SimpleHelp("OpenShift", "specifies the FeatureSet that is required to generate this XValidation rule.")),
	)
}

type FeatureSetEnum struct {
	FeatureSetNames []string `marker:"featureSet"`
	EnumValues      []string `marker:"enum"`
}

func (m FeatureSetEnum) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	if !RequiredFeatureSets.HasAny(m.FeatureSetNames...) {
		return nil
	}

	// TODO(directxman12): this is a bit hacky -- we should
	// probably support AnyType better + using the schema structure
	vals := make([]apiext.JSON, len(m.EnumValues))
	for i, val := range m.EnumValues {
		// TODO(directxman12): check actual type with schema type?
		// if we're expecting a string, marshal the string properly...
		// NB(directxman12): we use json.Marshal to ensure we handle JSON escaping properly
		valMarshalled, err := json.Marshal(val)
		if err != nil {
			return err
		}
		vals[i] = apiext.JSON{Raw: valMarshalled}
	}

	schema.Enum = vals
	return nil
}

type FeatureSetXValidation struct {
	FeatureSetNames []string `marker:"featureSet"`
	Rule            string
	Message         string `marker:",optional"`
}

func (m FeatureSetXValidation) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	if !RequiredFeatureSets.HasAny(m.FeatureSetNames...) {
		return nil
	}

	validation := XValidation{
		Rule:    m.Rule,
		Message: m.Message,
	}

	return validation.ApplyToSchema(schema)
}
