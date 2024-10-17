import Button from "../../ui/Button";
import Cross from "../../ui/Cross";
import Magnifyer from "../../ui/Magnifyer";
import Sun from "../../ui/Sun";

function AppFeatures() {
  return (
    <>
      <div className="flex gap-4 justify-center mb-4">
        <Button size="large">
          Odaljite
          <Magnifyer className="size-8" />
        </Button>
        <Button size="large">
          <Sun className="size-12" />
          {/* <Cross /> */}
        </Button>
        <Button size="large">
          Približite
          <Magnifyer className="size-8" />
        </Button>
      </div>

      <Button size="small" color="light" className="mx-auto mb-6">
        Sklonite dugmad
        <Cross />
      </Button>
    </>
  );
}

export default AppFeatures;
