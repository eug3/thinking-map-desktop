export default function SidebarLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen">
      <aside className="w-64 bg-gray-100 border-r">
        <div className="p-4">
          <h2 className="font-bold">ThinkingMap</h2>
        </div>
        <nav className="mt-4">
          <a href="/" className="block px-4 py-2 hover:bg-gray-200">Home</a>
          <a href="/workspace" className="block px-4 py-2 hover:bg-gray-200">Workspace</a>
        </nav>
      </aside>
      <main className="flex-1">{children}</main>
    </div>
  );
}
