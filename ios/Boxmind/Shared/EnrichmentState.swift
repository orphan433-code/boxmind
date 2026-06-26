import Foundation

enum EnrichmentState {
    private static let enrichmentWindow: TimeInterval = 5 * 60

    static func isEnriching(_ bookmark: Bookmark, now: Date = .now) -> Bool {
        if bookmark.enriched {
            return false
        }
        return now.timeIntervalSince(bookmark.createdAt) < enrichmentWindow
    }
}
