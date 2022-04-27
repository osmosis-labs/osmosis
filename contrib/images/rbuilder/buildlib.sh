#/bin/bash

f_make_release_tarball() {
    SOURCEDIST=${BASEDIR}/${APP}-${VERSION}.tar.gz

    git archive --format tar.gz --prefix "${APP}-${VERSION}/" -o "${SOURCEDIST}" HEAD

    l_tempdir="$(mktemp -d)"
    pushd "${l_tempdir}" >/dev/null
    tar xf "${SOURCEDIST}"
    rm "${SOURCEDIST}"
    find ${APP}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > "${SOURCEDIST}"
    popd >/dev/null
    rm -rf "${l_tempdir}"
}

f_setup_pristine_src_dir() {
    cd ${pristinesrcdir}
    tar --strip-components=1 -xf "${SOURCEDIST}"
    go mod download
}

f_exe_file_ext() {
   [ $(go env GOOS) = windows ] && printf '%s' '.exe' || printf ''
}

setup_build_env_for_platform() {
    local l_platform=$1

    g_old_GOOS="$(go env GOOS)"
    g_old_GOARCH="$(go env GOARCH)"
    g_old_OS_FILE_EXT="${OS_FILE_EXT}"

    go env -w GOOS="${l_platform%%/*}"
    go env -w GOARCH="${l_platform##*/}"
    OS_FILE_EXT="$(f_exe_file_ext)"
}

restore_build_env() {
    go env -w GOOS="${g_old_GOOS}"
    go env -w GOARCH="${g_old_GOARCH}"
    OS_FILE_EXT="${g_old_OS_FILE_EXT}"
}

generate_build_report() {
    local l_tempfile

    l_tempfile="$(mktemp)"

    pushd "${OUTDIR}" >/dev/null
    cat >>"${l_tempfile}" <<EOF
App: ${APP}
Version: ${VERSION}
Commit: ${COMMIT}
EOF
    echo 'Files:' >> "${l_tempfile}"
    md5sum * | sed 's/^/ /' >> "${l_tempfile}"
    echo 'Checksums-Sha256:' >> "${l_tempfile}"
    sha256sum * | sed 's/^/ /' >> "${l_tempfile}"
    mv "${l_tempfile}" build_report
    popd >/dev/null
}

[ "x${DEBUG}" = "x" ] || set -x

OS_FILE_EXT=''
BASEDIR="$(mktemp -d)"
OUTDIR=$HOME/artifacts
rm -rfv ${OUTDIR}/
mkdir -p ${OUTDIR}/
pristinesrcdir=${BASEDIR}/buildsources
mkdir -p ${pristinesrcdir}

# Make release tarball
f_make_release_tarball

# Extract release tarball and cache dependencies
f_setup_pristine_src_dir

# Move the release tarball to the out directory
mv ${SOURCEDIST} ${OUTDIR}/
