import "../src/styles.css";

export const metadata = {
  title: "Tecnomotos HB",
  description: "Sistema operativo del taller Tecnomotos HB"
};

export default function RootLayout({ children }) {
  return (
    <html lang="es">
      <body>{children}</body>
    </html>
  );
}
