import Header from "./ui/Header";
import "./index.css";
import Button from "./ui/Button";
import Container from "./ui/Container";
import ResultCard from "./ui/ResultCard";
import BottomLine from "./ui/BottomLine";

function App() {
  return (
    <>
      <Header />
      <Container>
        <div className="flex gap-4 justify-center mb-4">
          <Button size="large">Text</Button>
          <Button size="large">T</Button>
          <Button size="large">Text</Button>
        </div>
        <Button size="small" color="light" className="mx-auto mb-6">
          Sklonite dugmad
        </Button>

        <div className="flex gap-2 items-center mb-2 text-lg">
          <p className="text-xl text-main-800 font-semibold">
            Pitajte što god vas zanima 😀
          </p>
          <Button size="small" color="light">
            OČISTI
          </Button>
        </div>

        <textarea
          className="w-full bg-neutral-300 rounded-md px-2 py-2 text-2xl mb-4"
          rows={4}
          placeholder="Kako poslati sliku na fejzbuku"
        />

        <Button type="medium" className="w-full mb-8">
          Pretražite
        </Button>

        <ResultCard />

        <Button type="medium" color="light" className="w-full mb-8">
          Očisti rezultate
        </Button>
      </Container>

      <BottomLine />
    </>
  );
}

export default App;
