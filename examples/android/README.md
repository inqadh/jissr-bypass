# Android example

This is a minimal Android application that resolves Jissr Bypass from Maven
Central and performs a request away from the main thread.

Build with Gradle 9.4.1 or newer:

```sh
gradle :app:assembleDebug
```

To test another immutable package version:

```sh
gradle :app:assembleDebug -PjissrBypassVersion=0.1.0-rc.2
```

Install the generated debug APK on an Android API 26 or newer device. The
sample calls the public health endpoint and displays either the HTTP response
or a transport error.
