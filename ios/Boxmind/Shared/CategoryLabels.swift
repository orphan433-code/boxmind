import SwiftUI

enum CategoryLabels {
    static let order = [
        "movies",
        "music",
        "articles",
        "learning",
        "programming",
        "tools",
        "shopping",
        "jobs",
        "news",
        "design",
        "entertainment",
        "other"
    ]

    static let icons: [String: String] = [
        "movies": "play.rectangle.fill",
        "articles": "book.fill",
        "programming": "chevron.left.forwardslash.chevron.right",
        "shopping": "bag.fill",
        "jobs": "briefcase.fill",
        "learning": "graduationcap.fill",
        "music": "music.note",
        "news": "newspaper.fill",
        "design": "paintbrush.fill",
        "tools": "wrench.and.screwdriver.fill",
        "entertainment": "tv.fill",
        "other": "square.grid.2x2.fill"
    ]

    static let map: [String: String] = [
        "movies": "Смотреть",
        "articles": "Читать",
        "programming": "Полезное",
        "shopping": "Покупки",
        "jobs": "Работа",
        "learning": "Учиться",
        "music": "Слушать",
        "news": "Новости",
        "design": "Дизайн",
        "tools": "Полезное",
        "entertainment": "Смотреть",
        "other": "Другое"
    ]

    static func label(for category: String) -> String {
        map[category] ?? category
    }

    static func icon(for category: String) -> String {
        icons[category] ?? "folder.fill"
    }

    static let colors: [String: Color] = [
        "movies": .pink,
        "entertainment": .pink,
        "music": .purple,
        "articles": .orange,
        "news": .orange,
        "learning": .green,
        "programming": .blue,
        "tools": .blue,
        "shopping": .teal,
        "jobs": .cyan,
        "design": .indigo,
        "other": .gray
    ]

    static func color(for category: String) -> Color {
        colors[category] ?? .accentColor
    }

    static func sorted(_ categories: some Sequence<String>) -> [String] {
        categories.sorted { lhs, rhs in
            let lhsIndex = order.firstIndex(of: lhs) ?? Int.max
            let rhsIndex = order.firstIndex(of: rhs) ?? Int.max
            if lhsIndex != rhsIndex {
                return lhsIndex < rhsIndex
            }
            return label(for: lhs) < label(for: rhs)
        }
    }

    /// Display group key. Several raw categories (e.g. `movies` and `entertainment`)
    /// share one label, so we group by the visible label to avoid duplicate tiles.
    static func group(for category: String) -> String {
        label(for: category)
    }

    /// Unique display groups, ordered by the most prominent category in each group.
    static func sortedGroups(_ categories: some Sequence<String>) -> [String] {
        var bestOrderByGroup: [String: Int] = [:]
        for category in categories {
            let groupKey = group(for: category)
            let index = order.firstIndex(of: category) ?? Int.max
            if let existing = bestOrderByGroup[groupKey] {
                bestOrderByGroup[groupKey] = min(existing, index)
            } else {
                bestOrderByGroup[groupKey] = index
            }
        }

        return bestOrderByGroup.keys.sorted { lhs, rhs in
            let lhsIndex = bestOrderByGroup[lhs] ?? Int.max
            let rhsIndex = bestOrderByGroup[rhs] ?? Int.max
            if lhsIndex != rhsIndex {
                return lhsIndex < rhsIndex
            }
            return lhs < rhs
        }
    }

    /// Icon for a display group: the icon of its most prominent raw category.
    static func iconForGroup(_ group: String) -> String {
        let representative = order.first { label(for: $0) == group }
        return representative.map(icon(for:)) ?? "folder.fill"
    }
}
