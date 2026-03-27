import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MeetingListPage } from './pages/MeetingListPage'
import { CreateMeetingPage } from './pages/CreateMeetingPage'
import { MeetingDetailPage } from './pages/MeetingDetailPage'
import { ParticipantsPage } from './pages/ParticipantsPage'
import { DraftsPage } from './pages/DraftsPage'
import { LoginPage } from './pages/LoginPage'
import { authLogout } from './api/client'

const queryClient = new QueryClient()

function Layout({ children }: { children: React.ReactNode }) {
  const tabClass = ({ isActive }: { isActive: boolean }) =>
    isActive
      ? 'text-sm font-medium text-gray-900 border-b-2 border-gray-900 pb-3'
      : 'text-sm text-gray-500 hover:text-gray-700 pb-3'

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b sticky top-0 z-10">
        <div className="max-w-2xl mx-auto px-4">
          <div className="flex items-center justify-between pt-3 pb-0">
            <span className="font-semibold text-gray-900 text-sm">Редактор совещаний</span>
            <button onClick={authLogout} className="text-sm text-gray-400 hover:text-gray-600">Выйти</button>
          </div>
          <nav className="flex gap-6 mt-2">
            <NavLink to="/" end className={tabClass}>Совещания</NavLink>
            <NavLink to="/drafts" className={tabClass}>Черновики</NavLink>
            <NavLink to="/people" className={tabClass}>Реестр</NavLink>
          </nav>
        </div>
      </header>
      <main>{children}</main>
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/*"
            element={
              <Layout>
                <Routes>
                  <Route path="/" element={<MeetingListPage />} />
                  <Route path="/meetings/new" element={<CreateMeetingPage />} />
                  <Route path="/meetings/:id" element={<MeetingDetailPage />} />
                  <Route path="/drafts" element={<DraftsPage />} />
                  <Route path="/people" element={<ParticipantsPage />} />
                </Routes>
              </Layout>
            }
          />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
