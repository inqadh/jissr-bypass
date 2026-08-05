// swift-tools-version: 5.10
import PackageDescription

let package = Package(
    name: "JissrBypassAppleAdapter",
    platforms: [
        .iOS(.v15),
        .macOS(.v12),
    ],
    products: [
        .library(
            name: "JissrBypassURLSession",
            targets: ["JissrBypassURLSession"]
        ),
    ],
    targets: [
        .binaryTarget(
            name: "Jissrbypass",
            path: "Binaries/JissrBypass.xcframework"
        ),
        .target(
            name: "JissrBypassURLSession",
            dependencies: ["Jissrbypass"]
        ),
        .testTarget(
            name: "JissrBypassURLSessionTests",
            dependencies: ["JissrBypassURLSession"]
        ),
    ]
)
