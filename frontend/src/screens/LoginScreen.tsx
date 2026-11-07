const WhatsAppLogo = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 24 24" className={className} aria-label="WhatsApp" role="img">
    <path
      fill="#25D366"
      d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51a12.8 12.8 0 0 0-.57-.01c-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 0 1-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 0 1-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 0 1 2.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0 0 12.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 0 0 5.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 0 0-3.48-8.413"
    />
  </svg>
)

export function LoginScreen({
  canvasRef,
  status,
}: {
  canvasRef: React.RefObject<HTMLCanvasElement | null>
  status?: string
}) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-light-secondary dark:bg-dark-bg p-4">
      <div className="w-full max-w-225 h-130 bg-white dark:bg-gray-900 rounded-lg shadow-xl flex overflow-hidden border border-gray-100 dark:border-gray-800">
        {/* LEFT */}
        <div className="flex-1 p-12 flex flex-col justify-center">
          <div className="flex items-center gap-3 mb-8">
            <WhatsAppLogo className="w-10 h-10" />
            <h1 className="text-2xl font-light text-gray-700 dark:text-gray-200">
              Log in to WhatsApp
            </h1>
          </div>

          <ol className="list-decimal pl-6 space-y-4 text-gray-600 dark:text-gray-400 text-lg leading-relaxed">
            <li>Open WhatsApp on your phone</li>
            <li>
              Tap <b className="font-semibold text-gray-800 dark:text-gray-200">Menu</b> or{" "}
              <b className="font-semibold text-gray-800 dark:text-gray-200">Settings</b>
            </li>
            <li>
              Select{" "}
              <b className="font-semibold text-gray-800 dark:text-gray-200">Linked Devices</b>
            </li>
            <li>Point your phone at this screen</li>
          </ol>
        </div>

        {/* RIGHT */}
        <div className="flex-1 p-12 flex flex-col items-center justify-center border-l border-gray-100 dark:border-gray-800">
          <div className="bg-white p-2 rounded-lg shadow-sm mb-6">
            <canvas ref={canvasRef} className="size-66" />
          </div>

          <p className="text-gray-500 dark:text-gray-400 text-center mb-4">
            Scan this QR code with the WhatsApp app
          </p>

          <p className="text-sm font-medium text-light-primary dark:text-green-400 animate-pulse uppercase tracking-wide">
            {status === "timeout"
              ? "QR expired — refreshing…"
              : status === "refreshing"
                ? "Refreshing QR…"
                : status}
          </p>
        </div>
      </div>
    </div>
  )
}
