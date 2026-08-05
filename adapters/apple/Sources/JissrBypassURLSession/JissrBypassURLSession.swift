import Foundation
import JissrBypass

public enum JissrBypassURLSessionError: Error {
    case clientInitializationFailed
    case invalidProtectedHost
    case responseUnavailable
    case streamedRequestBodyUnsupported
}

/// Creates one centrally configured URLSession without changing URLRequest
/// call sites. Recreating the session discards its previous connection pool.
public enum JissrBypassURLSession {
    public static func patch(
        _ base: URLSessionConfiguration = .default,
        protectedHosts: [String]
    ) throws -> URLSession {
        let hosts = try Set(protectedHosts.map(canonicalHost))
        guard !hosts.isEmpty else {
            throw JissrBypassURLSessionError.invalidProtectedHost
        }

        guard let nativeClient = JissrbypassNewClient() else {
            throw JissrBypassURLSessionError.clientInitializationFailed
        }
        JissrBypassURLProtocol.configure(
            hosts: hosts,
            executor: { request in
                guard request.httpBodyStream == nil else {
                    throw JissrBypassURLSessionError.streamedRequestBodyUnsupported
                }
                let headers = try encodeHeaders(request.allHTTPHeaderFields ?? [:])
                let response = try nativeClient.execute(
                    request.httpMethod ?? "GET",
                    url: request.url?.absoluteString ?? "",
                    body: request.httpBody ?? Data(),
                    headersJSON: headers
                )
                guard let response else {
                    throw JissrBypassURLSessionError.responseUnavailable
                }
                return AdapterResponse(
                    statusCode: Int(response.getStatusCode()),
                    headersJSON: response.getHeadersJSON(),
                    body: response.getBody() ?? Data()
                )
            }
        )

        let configuration = base.copy() as! URLSessionConfiguration
        var protocolClasses = configuration.protocolClasses ?? []
        protocolClasses.removeAll { $0 == JissrBypassURLProtocol.self }
        protocolClasses.insert(JissrBypassURLProtocol.self, at: 0)
        configuration.protocolClasses = protocolClasses
        return URLSession(configuration: configuration)
    }
}

/// URL loading bridge used only by sessions returned from
/// JissrBypassURLSession.patch.
public final class JissrBypassURLProtocol: URLProtocol {
    private let stateLock = NSLock()
    private var stopped = false

    public override class func canInit(with request: URLRequest) -> Bool {
        guard
            request.url?.scheme?.lowercased() == "https",
            request.url?.port == nil || request.url?.port == 443,
            let host = request.url?.host,
            let canonical = try? canonicalHost(host)
        else {
            return false
        }
        return configuration()?.hosts.contains(canonical) == true
    }

    public override class func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    public override func startLoading() {
        guard let configuration = Self.configuration() else {
            fail(URLError(.badURL))
            return
        }
        let request = self.request
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let self, !self.isStopped else {
                return
            }
            do {
                let result = try configuration.executor(request)
                guard !self.isStopped, let url = request.url else {
                    return
                }
                let headerFields = try decodeHeaders(result.headersJSON)
                guard let response = HTTPURLResponse(
                    url: url,
                    statusCode: result.statusCode,
                    httpVersion: nil,
                    headerFields: headerFields
                ) else {
                    throw URLError(.cannotParseResponse)
                }
                guard !self.isStopped else {
                    return
                }
                self.client?.urlProtocol(
                    self,
                    didReceive: response,
                    cacheStoragePolicy: .notAllowed
                )
                self.client?.urlProtocol(self, didLoad: result.body)
                self.client?.urlProtocolDidFinishLoading(self)
            } catch {
                self.fail(error)
            }
        }
    }

    public override func stopLoading() {
        stateLock.lock()
        stopped = true
        stateLock.unlock()
    }

    private var isStopped: Bool {
        stateLock.lock()
        defer { stateLock.unlock() }
        return stopped
    }

    private func fail(_ error: Error) {
        guard !isStopped else {
            return
        }
        client?.urlProtocol(self, didFailWithError: error)
    }

    internal static func configure(
        hosts: Set<String>,
        executor: @escaping AdapterExecutor
    ) {
        registryLock.lock()
        activeConfiguration = AdapterConfiguration(
            hosts: hosts,
            executor: executor
        )
        registryLock.unlock()
    }

    fileprivate static func configuration() -> AdapterConfiguration? {
        registryLock.lock()
        defer { registryLock.unlock() }
        return activeConfiguration
    }

    private static let registryLock = NSLock()
    private static var activeConfiguration: AdapterConfiguration?
}

internal struct AdapterResponse {
    let statusCode: Int
    let headersJSON: String
    let body: Data
}

internal typealias AdapterExecutor = (URLRequest) throws -> AdapterResponse

fileprivate struct AdapterConfiguration {
    let hosts: Set<String>
    let executor: AdapterExecutor
}

fileprivate func canonicalHost(_ value: String) throws -> String {
    let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        .trimmingCharacters(in: CharacterSet(charactersIn: "."))
    guard
        !trimmed.isEmpty,
        let parsed = URL(string: "https://\(trimmed)"),
        let host = parsed.host?.lowercased(),
        !host.isEmpty,
        !host.contains(":"),
        !isIPv4Literal(host)
    else {
        throw JissrBypassURLSessionError.invalidProtectedHost
    }
    return host
}

fileprivate func isIPv4Literal(_ host: String) -> Bool {
    let labels = host.split(separator: ".", omittingEmptySubsequences: false)
    return labels.count == 4 && labels.allSatisfy { label in
        guard let value = Int(label) else {
            return false
        }
        return value >= 0 && value <= 255
    }
}

fileprivate func encodeHeaders(_ headers: [String: String]) throws -> String {
    let managed = Set(["connection", "content-length", "host", "transfer-encoding"])
    let filtered = headers.filter { !managed.contains($0.key.lowercased()) }
    let data = try JSONSerialization.data(withJSONObject: filtered)
    return String(decoding: data, as: UTF8.self)
}

fileprivate func decodeHeaders(_ encoded: String) throws -> [String: String] {
    let managed = Set(["connection", "content-length", "transfer-encoding"])
    let data = Data(encoded.utf8)
    let object = try JSONSerialization.jsonObject(with: data)
    guard let values = object as? [String: Any] else {
        throw URLError(.cannotParseResponse)
    }
    var result: [String: String] = [:]
    for (name, value) in values where !managed.contains(name.lowercased()) {
        if let strings = value as? [String] {
            result[name] = strings.joined(separator: ", ")
        } else if let string = value as? String {
            result[name] = string
        } else {
            throw URLError(.cannotParseResponse)
        }
    }
    return result
}
