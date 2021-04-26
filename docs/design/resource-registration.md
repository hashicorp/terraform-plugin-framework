# Registering Resources and Data Sources

There are two major steps to adding a resource or data source to a Terraform
provider. First, the resource or data source must be
[defined][structs-interfaces]. Second, the provider must be told about the
resource or data source. This design doc is concerned with that second bit, and
is meant to explore the different approaches we could take to allowing
resources and data sources to be registered with providers.

The end goal of each of these approaches is that the provider has a complete
list of the resources and data sources that are available through the provider,
and knows what name they'll be referred to with in HCL.

## Approaches

There are fundamentally two different approaches for this problem: the provider
can maintain a list of all the resources and data sources available, or the
resources and data sources can be held responsible for telling the provider
about themselves.

### Provider-Owned List

A provider-owned list (technically, map) is the approach that `helper/schema`
took:

```go
type Provider struct {
	Resources map[string]*Resource
	DataSources map[string]*Resource
}
```

then when defining your provider, typically in `provider.go`, you'd have
something like this:

```go
func New() *schema.Provider {
	// provider schema, etc. goes here
	Resources: map[string]*schema.Resource{
		"foo_compute_instance": resourceComputeInstance(),
		"foo_compute_disk": resourceComputeDisk(),
	},
	DataSources: map[string]*schema.Resource{
		"foo_compute_image": dataComputeImage(),
	},
	// provider_meta schema, etc. goes here
}
```

To add a new resource or data source, you just add another line to the map
definition.

### Resource-Owned Registration

Another approach is to have the resources and data sources own their own
registration, telling the provider about themselves:

```go
// inside resource_compute_instance.go
func init() {
	framework.RegisterResource("foo_compute_instance", resourceComputeInstance())
}
```

```go
// inside resource_compute_disk.go
func init() {
	framework.RegisterResource("foo_compute_disk", resourceComputeDisk())
}
```

```go
// inside data_compute_image.go
func init() {
	framework.RegisterDataSource("foo_compute_image", dataComputeImage())
}
```

## Trade-offs

There are a number of things to consider when evaluating these two approaches:

### Maintainability

One possible benefit of the resource-owned registration approach is that it
decentralizes all the code changes needed to add a resource. Meaning there's
not one chunk of lines that everyone trying to add a resource at the same time
is trying to modify. This removes a possibility of merge conflicts on large
providers that expand rapidly.

### Approachability

The resource-owned registration model is also (arguably) more approachable to
new developers; it means all the code changes required to add a resource can be
made in a single file, meaning provider developers don't need to remember to
modify multiple files, sometimes in multiple packages, to add a resource. The
concerns are all centralized in the resource itself.

On the other hand, the use of `init()` isn't incredibly obvious, and isn't the
most common pattern in Go programs, which may prove a barrier to entry.

### Global State

A major drawback of the resource-owned registration model is that it has to
rely on global mutable state. The resource needs to be registered to an
instance of a provider, and the provider instance it's being registered to
needs to exist when that registration happens. The only way around this is if a
global, mutable resource "registry" private state is kept, which can then be
assigned to provider instances as the provider instances are created. But even
then, the resource registry is still global mutable state.

This may have unanticipated consequences for the test framework and for any
other scenario where multiple provider servers could be running in parallel.

### Code Generation Friendliness

It is much easier to use code generation on a resource-owned registration
model, as the registration code is isolated per-resource. The init block just
needs to be added. Modifying a map is relatively more difficult, especially if
you want to order the resources and data sources in any fashion. Doubly so if
the values of the map _could_ be defined inline.

### Discovery

One drawback of the resource-owned registration model is that there is no
longer a single, central list of resources and data sources for the provider;
instead, the list is spread out over all the resource files. In theory, the
provider-owned list could have registrations spread out over a number of files,
as well, but inertia then is at least on the side of having them in a single,
neat list.

This only matters for crude, one-off scripts that are looking to e.g. count
resources in a provider without compiling it or executing the code. Arguably,
the same thing is still possible (perhaps easier?) with resource-owned
registration, although it needs to touch many more files.

## Notes

In theory, we don't _need_ to choose; if we choose to use the provider-owned
list method, provider devs can still choose to use the resource-owned
registration method:

```go
// in provider.go
var resources map[string]framework.Resource
var resourcesMu sync.Mutex

func registerResource(name string, r framework.Resource) {
	resourcesMu.Lock()
	defer resourcesMu.Unlock()

	if resources == nil {
		resources = map[string]framework.Resource{}
	}
	resources[name] = r
}

func New() framework.Provider {
	return framework.Provider{
		Resources: resources,
	}
}

// in resource_compute_instance.go
func init() {
	registerResource("foo_compute_instance", resourceComputeInstance())
}
```

This defaults to the safest option, and lets the provider decide to use the
less-safe (global mutable state) option. But that negates some of the benefits
of approachability for new developers.

[structs-interfaces]: https://github.com/hashicorp/terraform-plugin-framework/blob/main/docs/design/structs-interfaces.md
