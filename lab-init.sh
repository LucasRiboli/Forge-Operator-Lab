# organizar init do lab que inicia o processo dos pods
# simula o start do k8s kkkkk

incus init images:ubuntu/22.04 pod --config limits.memory=1GB


# stresse do pod, colocar em timer pra comecar

incus exec pod -- apt update
incus exec pod -- apt install -y stress-ng
incus exec pod -- stress-ng --vm 1 --vm-bytes 50% --timeout 300s