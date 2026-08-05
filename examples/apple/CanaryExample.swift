import Foundation
import JissrBypass

struct BypassResponse {
    let statusCode: Int
    let body: Data
    let headersJSON: String
}

func fetchWithJissrBypass() async throws -> BypassResponse {
    try await Task.detached(priority: .userInitiated) {
        guard let client = JissrbypassNewClient() else {
            throw NSError(domain: "JissrBypass", code: -1)
        }
        let response = try client.execute(
            "GET",
            url: "https://blocked.yusef.ren/v1/health",
            body: Data(),
            headersJSON: #"{"Accept":"application/json"}"#
        )
        return BypassResponse(
            statusCode: Int(response.getStatusCode()),
            body: response.getBody() ?? Data(),
            headersJSON: response.getHeadersJSON()
        )
    }.value
}
