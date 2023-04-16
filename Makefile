
envoy:
	export ENVOY_HOST=0.0.0.0 && \
		export ENVOY_PORT=10000 && \
		export UPLOAD_SERVICE_HOST=host.docker.internal && \
		export UPLOAD_SERVICE_PORT=9898 && \
		export AUDIO_SERVICE_HOST=host.docker.internal && \
		export AUDIO_SERVICE_PORT=9292 && \
		cat ./infra/envoy.tmpl.yaml | envsubst \$$ENVOY_HOST,\$$ENVOY_PORT,\$$UPLOAD_SERVICE_HOST,\$$UPLOAD_SERVICE_PORT,\$$AUDIO_SERVICE_HOST,\$$AUDIO_SERVICE_PORT \
		> ./infra/envoy.yaml

clean:
	docker-compose rm -f

.PHONY: envoy clean
