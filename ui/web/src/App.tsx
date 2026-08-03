import { BrowserRouter } from "react-router";
import { AppProviders } from "@/components/providers/app-providers";
import { AppRoutes } from "@/routes";

export default function App() {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <AppProviders>
        <AppRoutes />
      </AppProviders>
    </BrowserRouter>
  );
}
