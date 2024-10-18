import React from "react";
import Modal from "./Modal";
import { useState } from "react";

function ResultCard() {
  const [isOpenModal, setIsOpenModal] = useState(false);

  function handleOpenModal(param) {
    console.log(param);
    setIsOpenModal(param);
  }

  return (
    <div
      className="w-full rounded-md px-16 py-4  bg-neutral-100 mb-16"
      onClick={() => handleOpenModal(true)}
    >
      {isOpenModal && <Modal setShowModal={handleOpenModal} />}
      <p className="text-main-800 font-semibold">
        REZULTAT - Kliknite bilo gdje unutar sive površine za odlazak na
        stranicu
      </p>
      <h3 className="text-main-800 font-semibold text-3xl mb-4">
        Dijeljenje fotografija na Facebooku
      </h3>
      <p className="text-xl">
        Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi,
        reprehenderit molestias mollitia, distinctio voluptatum saepe facere
        laborum molestiae doloremque a asperiores minima, quidem dignissimos
        totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet
        consectetur adipisicing elit. Quis culpa provident amet inventore
        architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis,
        doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor
        sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit
        molestias mollitia, distinctio voluptatum saepe facere laborum molestiae
        doloremque a asperiores minima, quidem dignissimos totam aperiam amet
        repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing
        elit. Quis culpa provident amet inventore architecto vel tenetur placeat
        et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde!
        Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing
        elit. Quo eligendi, reprehenderit molestias mollitia, distinctio
        voluptatum saepe facere laborum molestiae doloremque a asperiores
        minima, quidem dignissimos totam aperiam amet repellendus error? Lorem,
        ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident
        amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque
        blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem
        ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi,
        reprehenderit molestias mollitia, distinctio voluptatum saepe facere
        laborum molestiae doloremque a asperiores minima, quidem dignissimos
        totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet
        consectetur adipisicing elit. Quis culpa provident amet inventore
        architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis,
        doloremque ullam velit fugit unde! Repudiandae, ipsam?
      </p>
    </div>
  );
}

export default ResultCard;
