import UIKit
import UniformTypeIdentifiers

/// Instant-save share extension: tap Boxmind in the share sheet, the bookmark is
/// posted to the API immediately and a small confirmation HUD is shown.
final class ShareViewController: UIViewController {
    private enum Outcome {
        case saved
        case duplicate
        case failed
    }

    private var didStart = false
    private let hud = SaveHUDView()

    override func viewDidLoad() {
        super.viewDidLoad()
        setupHUD()
    }

    override func viewDidAppear(_ animated: Bool) {
        super.viewDidAppear(animated)
        guard !didStart else { return }
        didStart = true

        showLoading()
        Task {
            let outcome = await save()
            await present(outcome)
        }
    }

    private func setupHUD() {
        hud.translatesAutoresizingMaskIntoConstraints = false
        hud.alpha = 0
        view.addSubview(hud)
        NSLayoutConstraint.activate([
            hud.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            hud.centerYAnchor.constraint(equalTo: view.centerYAnchor)
        ])
    }

    private func showLoading() {
        hud.configureLoading(title: "Сохраняем…")
        hud.transform = CGAffineTransform(scaleX: 0.8, y: 0.8)
        UIView.animate(withDuration: 0.25, delay: 0, usingSpringWithDamping: 0.7, initialSpringVelocity: 0.5) {
            self.hud.alpha = 1
            self.hud.transform = .identity
        }
    }

    private func save() async -> Outcome {
        guard let token = AppGroupStore.accessToken,
              let baseURL = AppGroupStore.apiBaseURL,
              let url = await firstSharedURL() else {
            return .failed
        }

        return await postBookmark(urlString: url.absoluteString, token: token, baseURL: baseURL)
    }

    @MainActor
    private func present(_ outcome: Outcome) async {
        switch outcome {
        case .saved:
            hud.configure(symbol: "checkmark.circle.fill", title: "Сохранено", tint: .systemGreen)
            UINotificationFeedbackGenerator().notificationOccurred(.success)
        case .duplicate:
            hud.configure(symbol: "bookmark.circle.fill", title: "Уже сохранено", tint: .systemBlue)
            UINotificationFeedbackGenerator().notificationOccurred(.warning)
        case .failed:
            hud.configure(symbol: "exclamationmark.circle.fill", title: "Не удалось", tint: .systemRed)
            UINotificationFeedbackGenerator().notificationOccurred(.error)
        }

        try? await Task.sleep(for: .milliseconds(750))

        UIView.animate(withDuration: 0.2) {
            self.hud.alpha = 0
            self.view.alpha = 0
        } completion: { _ in
            self.finish()
        }
    }

    // MARK: - Networking

    private func postBookmark(urlString: String, token: String, baseURL: String) async -> Outcome {
        guard let endpoint = URL(string: baseURL.hasSuffix("/") ? "\(baseURL)bookmarks" : "\(baseURL)/bookmarks") else {
            return .failed
        }

        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.httpBody = try? JSONSerialization.data(withJSONObject: ["url": urlString])

        guard let (_, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse else {
            return .failed
        }

        switch http.statusCode {
        case 200 ... 299:
            return .saved
        case 409:
            return .duplicate
        default:
            return .failed
        }
    }

    // MARK: - Shared item parsing

    private func firstSharedURL() async -> URL? {
        guard let item = extensionContext?.inputItems.first as? NSExtensionItem,
              let providers = item.attachments else {
            return nil
        }

        var textCandidates: [String] = []

        for provider in providers {
            if provider.hasItemConformingToTypeIdentifier(UTType.url.identifier),
               let url = await loadURL(from: provider, typeIdentifier: UTType.url.identifier),
               isWebURL(url) {
                return url
            }

            if provider.hasItemConformingToTypeIdentifier(UTType.plainText.identifier),
               let text = await loadString(from: provider, typeIdentifier: UTType.plainText.identifier) {
                textCandidates.append(text)
            }

            if provider.hasItemConformingToTypeIdentifier(UTType.propertyList.identifier),
               let plist = await loadPropertyList(from: provider),
               let url = urlFromPropertyList(plist) {
                return url
            }
        }

        for text in textCandidates {
            if let url = URL(string: text.trimmingCharacters(in: .whitespacesAndNewlines)), isWebURL(url) {
                return url
            }
            if let url = extractURL(from: text) {
                return url
            }
        }

        return nil
    }

    private func isWebURL(_ url: URL) -> Bool {
        guard let scheme = url.scheme?.lowercased() else { return false }
        return scheme == "http" || scheme == "https"
    }

    private func extractURL(from text: String) -> URL? {
        let pattern = #"https?://[^\s<>"']+"#
        guard let range = text.range(of: pattern, options: .regularExpression) else {
            return nil
        }
        let raw = String(text[range])
        return URL(string: raw)
    }

    private func urlFromPropertyList(_ plist: [String: Any]) -> URL? {
        let keys = ["URL", "url", "NSExtensionJavaScriptPreprocessingResults"]
        for key in keys {
            if let value = plist[key] as? String, let url = URL(string: value), isWebURL(url) {
                return url
            }
            if let nested = plist[key] as? [String: Any],
               let urlString = nested["URL"] as? String ?? nested["url"] as? String,
               let url = URL(string: urlString), isWebURL(url) {
                return url
            }
        }
        return nil
    }

    private func loadURL(from provider: NSItemProvider, typeIdentifier: String) async -> URL? {
        await withCheckedContinuation { continuation in
            provider.loadItem(forTypeIdentifier: typeIdentifier, options: nil) { item, _ in
                if let url = item as? URL {
                    continuation.resume(returning: url)
                } else if let text = item as? String {
                    continuation.resume(returning: URL(string: text))
                } else {
                    continuation.resume(returning: nil)
                }
            }
        }
    }

    private func loadString(from provider: NSItemProvider, typeIdentifier: String) async -> String? {
        await withCheckedContinuation { continuation in
            provider.loadItem(forTypeIdentifier: typeIdentifier, options: nil) { item, _ in
                continuation.resume(returning: item as? String)
            }
        }
    }

    private func loadPropertyList(from provider: NSItemProvider) async -> [String: Any]? {
        await withCheckedContinuation { continuation in
            provider.loadItem(forTypeIdentifier: UTType.propertyList.identifier, options: nil) { item, _ in
                continuation.resume(returning: item as? [String: Any])
            }
        }
    }

    private func finish() {
        extensionContext?.completeRequest(returningItems: nil)
    }
}

/// Small rounded confirmation card with an SF Symbol and a title, à la system HUDs.
private final class SaveHUDView: UIView {
    private let iconView = UIImageView()
    private let spinner = UIActivityIndicatorView(style: .large)
    private let titleLabel = UILabel()

    override init(frame: CGRect) {
        super.init(frame: frame)
        setup()
    }

    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setup()
    }

    private func setup() {
        let blur = UIVisualEffectView(effect: UIBlurEffect(style: .systemMaterial))
        blur.translatesAutoresizingMaskIntoConstraints = false
        blur.layer.cornerRadius = 22
        blur.layer.cornerCurve = .continuous
        blur.clipsToBounds = true
        addSubview(blur)

        iconView.translatesAutoresizingMaskIntoConstraints = false
        iconView.contentMode = .scaleAspectFit
        iconView.preferredSymbolConfiguration = UIImage.SymbolConfiguration(pointSize: 44, weight: .semibold)

        titleLabel.translatesAutoresizingMaskIntoConstraints = false
        titleLabel.font = .systemFont(ofSize: 17, weight: .semibold)
        titleLabel.textColor = .label
        titleLabel.textAlignment = .center

        spinner.translatesAutoresizingMaskIntoConstraints = false
        spinner.hidesWhenStopped = true

        let iconContainer = UIView()
        iconContainer.translatesAutoresizingMaskIntoConstraints = false
        iconContainer.addSubview(iconView)
        iconContainer.addSubview(spinner)

        let stack = UIStackView(arrangedSubviews: [iconContainer, titleLabel])
        stack.axis = .vertical
        stack.alignment = .center
        stack.spacing = 12
        stack.translatesAutoresizingMaskIntoConstraints = false
        blur.contentView.addSubview(stack)

        NSLayoutConstraint.activate([
            iconContainer.widthAnchor.constraint(equalToConstant: 48),
            iconContainer.heightAnchor.constraint(equalToConstant: 48),
            iconView.centerXAnchor.constraint(equalTo: iconContainer.centerXAnchor),
            iconView.centerYAnchor.constraint(equalTo: iconContainer.centerYAnchor),
            spinner.centerXAnchor.constraint(equalTo: iconContainer.centerXAnchor),
            spinner.centerYAnchor.constraint(equalTo: iconContainer.centerYAnchor)
        ])

        NSLayoutConstraint.activate([
            blur.topAnchor.constraint(equalTo: topAnchor),
            blur.bottomAnchor.constraint(equalTo: bottomAnchor),
            blur.leadingAnchor.constraint(equalTo: leadingAnchor),
            blur.trailingAnchor.constraint(equalTo: trailingAnchor),

            widthAnchor.constraint(greaterThanOrEqualToConstant: 160),
            heightAnchor.constraint(greaterThanOrEqualToConstant: 120),

            stack.centerXAnchor.constraint(equalTo: blur.contentView.centerXAnchor),
            stack.centerYAnchor.constraint(equalTo: blur.contentView.centerYAnchor),
            stack.leadingAnchor.constraint(greaterThanOrEqualTo: blur.contentView.leadingAnchor, constant: 20),
            stack.trailingAnchor.constraint(lessThanOrEqualTo: blur.contentView.trailingAnchor, constant: -20),
            stack.topAnchor.constraint(greaterThanOrEqualTo: blur.contentView.topAnchor, constant: 20),
            stack.bottomAnchor.constraint(lessThanOrEqualTo: blur.contentView.bottomAnchor, constant: -20)
        ])
    }

    func configureLoading(title: String) {
        iconView.isHidden = true
        iconView.image = nil
        spinner.startAnimating()
        titleLabel.text = title
    }

    func configure(symbol: String, title: String, tint: UIColor) {
        spinner.stopAnimating()
        iconView.isHidden = false
        iconView.image = UIImage(systemName: symbol)
        iconView.tintColor = tint
        titleLabel.text = title
    }
}
