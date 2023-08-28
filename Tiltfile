allow_k8s_contexts('kind-nuc-dev')

operator = [
  'neosync-operator',
]
db = [
  'postgresql',
]
backend = [
  'neosync-api',
]
frontend = [
  'neosync-app',
]

groups = {
  'operator': operator,
  'backend': backend + db,
  'frontend': backend + db + frontend,
}
config.define_string_list("to-run", args=True)
cfg = config.parse()
resources = []
for arg in cfg.get('to-run', []):
  if arg in groups:
    resources += groups[arg]
  else:
    # also support specifying individual services instead of groups, e.g. `tilt up a b d`
    resources.append(arg)
config.set_enabled_resources(resources)

load_dynamic('k8s-operator/Tiltfile')
load_dynamic('backend/Tiltfile')
load_dynamic('frontend/Tiltfile')
