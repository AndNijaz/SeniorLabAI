import { useTheme } from "../features/theme-select/ThemeProvider";

function Header() {
  const { theme } = useTheme();

  return (
    <header className="w-full bg-neutral-200 py-4 flex justify-center dark:bg-gray-800 text-[#b0b0b0ea]">
      <img
        src={`/public/${
          theme === "light" ? "SeniorLabLogo.svg" : "SeniorLabLogoBlue.svg"
        }`}
        alt="SeniorLab Logo"
        className="w-40 text-white"
      />
    </header>
  );
}

export default Header;
