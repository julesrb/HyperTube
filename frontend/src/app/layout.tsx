import Navabar from "@/components/Navbar";

export default function RootLayout({children, }: { children: React.ReactNode; }) {
    return (
        <html>
        <body>

        <Navabar />

        {children}

        </body>
        </html>
    );
}
