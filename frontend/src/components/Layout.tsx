import { Link, Outlet, useLocation } from 'react-router-dom';
import SyncButton from './SyncButton';
import WorkflowTrigger from './WorkflowTrigger';

interface LayoutProps {
  onSyncComplete?: () => void;
}

export default function Layout({ onSyncComplete }: LayoutProps) {
  const location = useLocation();
  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">0xNetworth</h1>
              <p className="mt-2 text-gray-600">Investment Tracking Dashboard</p>
            </div>
            <div className="flex items-center gap-4">
              <WorkflowTrigger onExecutionComplete={(execution) => {
                console.log('Workflow execution started:', execution);
              }} />
              <SyncButton onSyncComplete={onSyncComplete} />
            </div>
          </div>
          
          {/* Navigation */}
          <nav className="mt-6 flex gap-4">
            <Link
              to="/"
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                isActive('/')
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              Dashboard
            </Link>
            <Link
              to="/workflows"
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                isActive('/workflows')
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              Workflows
            </Link>
            <Link
              to="/sources"
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                isActive('/sources')
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              YouTube Sources
            </Link>
          </nav>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}

