allow_k8s_contexts('kind-nuc-dev')

backend = [
  'neosync-api',
]
frontend = [
  'neosync-app',
]

groups = {
  'backend': backend,
  'frontend': backend + frontend,
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

load_dynamic('tilt/temporal/Tiltfile')
load_dynamic('backend/Tiltfile')
load_dynamic('frontend/Tiltfile')
