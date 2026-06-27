import SwiftUI

struct RootView: View {
    @Environment(SessionStore.self) private var session
    @AppStorage("hasCompletedOnboarding") private var hasCompletedOnboarding = false

    var body: some View {
        Group {
            if session.isAuthenticated {
                if hasCompletedOnboarding {
                    MainTabView()
                } else {
                    OnboardingView {
                        hasCompletedOnboarding = true
                    }
                }
            } else {
                LoginView()
            }
        }
        .animation(.easeInOut(duration: 0.2), value: session.isAuthenticated)
        .animation(.easeInOut(duration: 0.2), value: hasCompletedOnboarding)
    }
}

#Preview {
    RootView()
        .environment(SessionStore())
}
