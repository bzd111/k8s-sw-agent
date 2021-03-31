FROM scratch

ARG ARCH

COPY ./dist/pod-admission-webhooklinux_${ARCH} /usr/bin/pod-admission-webhook
ENTRYPOINT [ "/usr/bin/pod-admission-webhook"]
