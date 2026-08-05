#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <version> <aar> <sources-jar> <output-zip>" >&2
}

if [[ $# -ne 4 ]]; then
  usage
  exit 2
fi

version="$1"
aar_path="$2"
sources_path="$3"
output_path="$4"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] ||
   [[ "$version" == *-SNAPSHOT ]]; then
  echo "Invalid immutable Maven version: $version" >&2
  exit 2
fi
if [[ ! -f "$aar_path" ]]; then
  echo "Android AAR not found: $aar_path" >&2
  exit 2
fi
if [[ ! -f "$sources_path" ]]; then
  echo "Sources JAR not found: $sources_path" >&2
  exit 2
fi
if [[ -z "${MAVEN_SIGNING_KEY_B64:-}" ]]; then
  echo "MAVEN_SIGNING_KEY_B64 is required." >&2
  exit 2
fi

for command_name in base64 gpg jar javadoc md5sum sha1sum sha256sum sha512sum unzip zip; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command is unavailable: $command_name" >&2
    exit 2
  fi
done

group_id="org.jissr"
artifact_id="jissr-bypass-android"
group_path="${group_id//./\/}"
artifact_prefix="${artifact_id}-${version}"
output_path="$(realpath -m "$output_path")"
work_dir="$(mktemp -d)"
trap 'rm -rf -- "$work_dir"' EXIT

release_dir="$work_dir/repository/$group_path/$artifact_id/$version"
mkdir -p "$release_dir" "$(dirname "$output_path")"

aar_name="$artifact_prefix.aar"
sources_name="$artifact_prefix-sources.jar"
javadoc_name="$artifact_prefix-javadoc.jar"
pom_name="$artifact_prefix.pom"

cp "$aar_path" "$release_dir/$aar_name"
cp "$sources_path" "$release_dir/$sources_name"

mkdir -p "$work_dir/javadoc" "$work_dir/sources" "$work_dir/aar"
unzip -q "$sources_path" -d "$work_dir/sources"
unzip -q "$aar_path" classes.jar -d "$work_dir/aar"
mapfile -t java_sources < <(
  find "$work_dir/sources/jissrbypass" -type f -name '*.java' | sort
)
if [[ ${#java_sources[@]} -eq 0 ]]; then
  echo "No public jissrbypass Java sources were found in: $sources_path" >&2
  exit 2
fi
javadoc \
  -quiet \
  -Xdoclint:none \
  -encoding UTF-8 \
  -docencoding UTF-8 \
  -charset UTF-8 \
  -windowtitle "Jissr Bypass Android API" \
  -doctitle "Jissr Bypass Android API" \
  -classpath "$work_dir/aar/classes.jar" \
  -d "$work_dir/javadoc" \
  "${java_sources[@]}"
cp docs/maven/README.md "$work_dir/javadoc/README.md"
(
  cd "$work_dir/javadoc"
  jar --create --file "$release_dir/$javadoc_name" .
)

cat >"$release_dir/$pom_name" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>$group_id</groupId>
  <artifactId>$artifact_id</artifactId>
  <version>$version</version>
  <packaging>aar</packaging>
  <name>Jissr Bypass for Android</name>
  <description>Native ECH HTTPS transport for Android applications.</description>
  <url>https://github.com/inqadh/jissr-bypass</url>
  <organization>
    <name>المجلس الوطني للإنقاذ</name>
    <url>https://www.inqadh.com/</url>
  </organization>
  <licenses>
    <license>
      <name>Apache License, Version 2.0</name>
      <url>https://www.apache.org/licenses/LICENSE-2.0.txt</url>
      <distribution>repo</distribution>
    </license>
  </licenses>
  <developers>
    <developer>
      <id>inqadh</id>
      <name>المجلس الوطني للإنقاذ</name>
      <email>releases@inqadh.com</email>
      <organization>المجلس الوطني للإنقاذ</organization>
      <organizationUrl>https://www.inqadh.com/</organizationUrl>
    </developer>
  </developers>
  <scm>
    <connection>scm:git:https://github.com/inqadh/jissr-bypass.git</connection>
    <developerConnection>scm:git:ssh://git@github.com/inqadh/jissr-bypass.git</developerConnection>
    <url>https://github.com/inqadh/jissr-bypass</url>
    <tag>v$version</tag>
  </scm>
  <issueManagement>
    <system>GitHub Issues</system>
    <url>https://github.com/inqadh/jissr-bypass/issues</url>
  </issueManagement>
</project>
EOF

export GNUPGHOME="$work_dir/gnupg"
mkdir -m 700 "$GNUPGHOME"
printf '%s' "$MAVEN_SIGNING_KEY_B64" | base64 --decode >"$work_dir/signing-key.asc"
gpg --batch --import "$work_dir/signing-key.asc" >/dev/null 2>&1
signing_fingerprint="$(
  gpg --batch --with-colons --list-secret-keys |
    awk -F: '$1 == "fpr" { print $10; exit }'
)"
if [[ -z "$signing_fingerprint" ]]; then
  echo "No signing key was imported." >&2
  exit 2
fi

primary_files=(
  "$release_dir/$aar_name"
  "$release_dir/$sources_name"
  "$release_dir/$javadoc_name"
  "$release_dir/$pom_name"
)

for artifact in "${primary_files[@]}"; do
  gpg_args=(
    --batch
    --yes
    --armor
    --detach-sign
    --local-user "$signing_fingerprint"
    --output "$artifact.asc"
  )
  if [[ -n "${MAVEN_SIGNING_PASSWORD:-}" ]]; then
    gpg_args+=(--pinentry-mode loopback --passphrase "$MAVEN_SIGNING_PASSWORD")
  fi
  gpg "${gpg_args[@]}" "$artifact"

  md5sum "$artifact" | awk '{print $1}' >"$artifact.md5"
  sha1sum "$artifact" | awk '{print $1}' >"$artifact.sha1"
  sha256sum "$artifact" | awk '{print $1}' >"$artifact.sha256"
  sha512sum "$artifact" | awk '{print $1}' >"$artifact.sha512"
done

(
  cd "$work_dir/repository"
  zip -q -r "$output_path" .
)

echo "Maven Central bundle: $output_path"
echo "Coordinate: $group_id:$artifact_id:$version"
echo "Signing key: $signing_fingerprint"
