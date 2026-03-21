import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MeetingListPage } from './pages/MeetingListPage'
import { CreateMeetingPage } from './pages/CreateMeetingPage'
import { MeetingDetailPage } from './pages/MeetingDetailPage'
import { ParticipantsPage } from './pages/ParticipantsPage'

const queryClient = new QueryClient()

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b sticky top-0 z-10">
        <div className="max-w-2xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="font-semibold text-gray-900 text-sm">Редактор совещаний</Link>
          <Link to="/people" className="text-sm text-gray-500 hover:text-gray-700">Участники</Link>
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
        <Layout>
          <Routes>
            <Route path="/" element={<MeetingListPage />} />
            <Route path="/meetings/new" element={<CreateMeetingPage />} />
            <Route path="/meetings/:id" element={<MeetingDetailPage />} />
            <Route path="/people" element={<ParticipantsPage />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
