import "./index.css";
import Container from "./ui/Container";
import ResultCard from "./features/Search/ResultCard";
import Search from "./features/Search/Search";
import AppFeatures from "./features/app-features/AppFeatures";
import { ClearSearchButton, useSearch } from "./features/Search/SearchProvider";
import AppLayout from "./ui/AppLayout";

function App() {
  const { search } = useSearch();

  console.log(JSON.stringify(import.meta.env.VITE_OPEN_AI_ENDPOINT));

  return (
    <AppLayout>
      <Container className="pt-8">
        {/* Features Buttons (zoom in, zoom out, theme); Hide Buttons Button */}
        <AppFeatures />

        {/* Text Area Label; Text Area; Clear Button */}
        <Search />

        {/* Resault Caert with its features */}
        {search && <ResultCard />}

        {search && <ClearSearchButton />}
      </Container>
    </AppLayout>
  );
}
//   return (
//     <>
//       <Header />
//       <Container>

//         {/* Features Buttons (zoom in, zoom out, theme); Hide Buttons Button */}
//         <AppFeatures />

//         {/* Text Area Label; Text Area; Clear Button */}
//         <Search />

//         {/* Resault Caert with its features */}
//         {result && <ResultCard />}

//         {result && (
//           <Button type="medium" color="light" className="w-full mb-32">
//             Očisti rezultate
//           </Button>
//         )}
//       </Container>

//       <Footer />
//     </>
//   );
// }

export default App;
