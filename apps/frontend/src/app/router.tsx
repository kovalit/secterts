import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '../components/layout/AppLayout'
import { ProtectedRoute } from '../features/auth/ProtectedRoute'
import { LoginPage } from '../features/auth/LoginPage'
import { RegisterPage } from '../features/auth/RegisterPage'
import { VerifyEmailCodePage } from '../features/auth/VerifyEmailCodePage'
import { PasswordsPage } from '../features/passwords/PasswordsPage'
import { CompaniesPage } from '../features/companies/CompaniesPage'
import { SecretNotesPage } from '../features/secret-notes/SecretNotesPage'
import { ProjectsPage } from '../features/app-secrets/ProjectsPage'
import { ProjectDetailPage } from '../features/app-secrets/ProjectDetailPage'
import { ExtensionTokensPage } from '../features/settings/ExtensionTokensPage'
import { SecurityPage } from '../features/settings/SecurityPage'
import { AuditPage } from '../features/settings/AuditPage'

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/verify-email-code', element: <VerifyEmailCodePage /> },
  {
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { path: '/', element: <Navigate to="/passwords" replace /> },
      { path: '/passwords', element: <PasswordsPage /> },
      { path: '/companies', element: <CompaniesPage /> },
      { path: '/secret-notes', element: <SecretNotesPage /> },
      { path: '/app-secrets', element: <ProjectsPage /> },
      { path: '/app-secrets/:projectId', element: <ProjectDetailPage /> },
      { path: '/settings/security', element: <SecurityPage /> },
      { path: '/settings/extension', element: <ExtensionTokensPage /> },
      { path: '/settings/audit', element: <AuditPage /> },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
])
