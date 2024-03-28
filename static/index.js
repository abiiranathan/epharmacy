const salesQueue = document.getElementById("salesQueue");
const tbody = document.getElementById("productsTable");
const barcodeInput = document.getElementById("barcode");
const createTransaction = document.querySelector(".create-transaction");
const grandTotal = document.getElementById("grand_total");

const numberFormatter = Intl.NumberFormat("en-GB", {
  currency: "UGX",
  maximumFractionDigits: 2,
  minimumFractionDigits: 2,
});

barcodeInput.focus();

function addProductToQueue(product) {
  const tr = document.createElement("tr");

  tr.innerHTML = `
      <td class="hidden queue-id">${product.id}</td>
      <td>${product.generic_name}</td>
      <td>${product.brand_name}</td>
      <td class="queue-selling_price">${product.selling_price.toFixed(2)}</td>
      <td class="queue-quantity" style="background-color: lightgreen; font-size:16px;" contenteditable>${
        product.quantity
      }</td>
      <td class="queue-subtotal">${(product.selling_price * product.quantity).toFixed(2)}</td>
      <td>
        <button class="button remove-button" data-id="${product.id}">Remove</button>
      </td>
    `;
  salesQueue.appendChild(tr);
}

function setProducts(products) {
  tbody.innerHTML = "";

  products.forEach((product) => {
    const tr = document.createElement("tr");
    if (product.quantity == 0) {
      tr.style.backgroundColor = "lightcoral";
    }

    tr.innerHTML = `
        <td class="GenericName">${product.generic_name}</td>
        <td class="BrandName">${product.brand_name}</td>
        <td class="Quantity" id="quantity-${product.id}">${product.quantity}</td>
        <td class="SellingPrice">${product.selling_price}</td>
        <td class="ExpiryDates">
            <div class="flex gap-4 items-center">
              ${product.expiry_dates
                .map(
                  (date) =>
                    `<p class="ExpiryDate">${new Date(date).toLocaleDateString(undefined, {
                      year: "numeric",
                      month: "long",
                    })}</p>`,
                )
                .join("")}
            </div>
        </td>
        <td>
          <button ${product.quantity == 0 ? "disabled" : ""} class="button add-button" data-id="${product.id}">Add</button>
        </td>
      `;
    tbody.appendChild(tr);
  });
}

function computeGrandTotal() {
  const subtotalElements = salesQueue.querySelectorAll(".queue-subtotal");
  const subtotals = Array.from(subtotalElements).map((el) => {
    return parseFloat(el.textContent.trim());
  });

  const sum = subtotals.reduce((prev, curr) => prev + curr, 0);

  grandTotal.innerText = numberFormatter.format(sum);
}

function resetGrandTotal() {
  grandTotal.innerHTML = "0.00";
}

function maxQuantityExceeded(product) {
  const qtyElement = document.getElementById("quantity-" + product.id);
  if (!qtyElement) return false;

  const availableQty = parseFloat(qtyElement.innerText.trim());
  const queueItems = salesQueue.querySelectorAll(".queue-id");
  let totalQuantity = 0;

  for (const item of queueItems) {
    if (parseInt(item.textContent.trim()) === product.id) {
      const tr = item.closest("tr");
      const quantity = parseInt(tr.querySelector(".queue-quantity").textContent.trim());
      totalQuantity += quantity;
      break;
    }
  }
  return totalQuantity >= availableQty;
}

function addProductOrUpdate(product) {
  if (maxQuantityExceeded(product)) {
    console.log("Maximum quantity exceeded for this product!");
    return;
  }

  const queueItems = salesQueue.querySelectorAll(".queue-id");
  for (const item of queueItems) {
    if (parseInt(item.textContent.trim()) === product.id) {
      const tr = item.closest("tr");
      const quantity = parseInt(tr.querySelector(".queue-quantity").textContent.trim());
      tr.querySelector(".queue-quantity").textContent = quantity + 1;
      tr.querySelector(".queue-subtotal").textContent = (quantity + 1) * product.selling_price;
      return;
    }
  }
  addProductToQueue(product);
}

// event delegation
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("add-button")) {
    const tr = e.target.closest("tr");
    if (!tr) return;
    const id = e.target.getAttribute("data-id");
    const generic_name = tr.querySelector(".GenericName").textContent.trim();
    const brand_name = tr.querySelector(".BrandName").textContent.trim();
    const selling_price = tr.querySelector(".SellingPrice").textContent.trim();
    const expiry_dates = Array.from(tr.querySelectorAll(".ExpiryDate")).map((date) =>
      date.textContent.trim(),
    );
    const quantity = 1;

    console.log(expiry_dates);

    addProductOrUpdate({
      id: parseInt(id),
      generic_name,
      brand_name,
      selling_price: parseFloat(selling_price),
      expiry_dates,
      quantity,
    });

    // clear the search input
    document.querySelector(".search_product").value = "";

    // focus on the barcode input
    barcodeInput.focus();

    // compute grand total
    computeGrandTotal();
  }
});

// event delegation
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("remove-button")) {
    const tr = e.target.closest("tr");
    if (!tr) return;
    tr.remove();

    // compute grand total
    computeGrandTotal();
  }
});

// event delegation when quantity is changed
document.addEventListener("input", (e) => {
  if (e.target.classList.contains("queue-quantity")) {
    const tr = e.target.closest("tr");
    if (!tr) return;

    let quantity = parseInt(e.target.textContent) || 0;
    const id = tr.querySelector(".queue-id").textContent.trim();
    const availableQtyElem = document.getElementById("quantity-" + id);
    if (availableQtyElem) {
      const availableQty = parseFloat(availableQtyElem.innerText.trim());
      if (quantity > availableQty) {
        quantity = availableQty;
        e.target.textContent = quantity;
        alert("Insufficient quantity in stock. Available quantity: " + availableQty);
      }
    }

    const sellingPrice = parseFloat(tr.querySelector(".queue-selling_price").textContent);
    tr.querySelector(".queue-subtotal").textContent = (quantity * sellingPrice).toFixed(2);

    // compute grand total
    computeGrandTotal();
  } else if (e.target.name === "search_product") {
    // Searching products
    const search = e.target.value.trim();

    const url = `/products/search?name=${search}`;
    fetch(url)
      .then(async (res) => {
        if (!res.ok) {
          throw await res.json();
        }
        return res.json();
      })
      .then((data) => {
        // set the products.
        setProducts(data);
      })
      .catch((error) => console.error(error));
  } else if (e.target.name === "barcode") {
    // Searching products and automatically adding to the sales queue a quantity of 1
    const search = e.target.value.trim();
    if (search == "") {
      return;
    }

    const url = `/products/search/barcode/${search}`;
    fetch(url)
      .then((res) => {
        if (!res.ok) {
          throw new Error("Product not found!");
        }
        return res.json();
      })
      .then((data) => {
        if (data.quantity == 0) {
          alert(data.generic_name + "is out of stock!");
        } else {
          addProductOrUpdate({ ...data, quantity: 1 });

          // compute grand total
          computeGrandTotal();
        }

        // clear the search input
        e.target.value = "";
      })
      .catch((error) => console.error(error));
  }
});

// event delegation
createTransaction.addEventListener("click", async (e) => {
  const url = e.currentTarget.getAttribute("data-url");
  const method = e.currentTarget.getAttribute("data-method");

  const invalidProducts = Array.from(salesQueue.children).some((tr) => {
    const quantity = parseInt(tr.querySelector(".queue-quantity").textContent.trim()) || 0;
    return isNaN(quantity) || quantity <= 0;
  });

  if (invalidProducts) {
    alert("Invalid quantity for some products!");
    return;
  }

  const products = Array.from(salesQueue.children).map((tr) => {
    const id = tr.querySelector(".queue-id").textContent.trim();
    const SellingPrice = tr.querySelector(".queue-selling_price").textContent.trim();
    const Quantity = tr.querySelector(".queue-quantity").textContent.trim();

    return {
      id: parseInt(id),
      selling_price: parseFloat(SellingPrice),
      quantity: parseInt(Quantity),
    };
  });

  if (products.length == 0) {
    alert("No products in the sales queue or quantity is 0!");
    return;
  }

  const response = await fetch(url, {
    method,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ products }),
  });

  try {
    const data = await response.json();
    if (response.ok) {
      salesQueue.innerHTML = "";
      decrementQuantities(products);

      // compute grand total
      resetGrandTotal();
    } else {
      alert(data.error || "Insufficient quantity in stock!");
    }
  } catch (error) {
    console.error(error);
    alert("An error occurred");
  }
});

function decrementQuantities(products) {
  for (const prod of products) {
    const qtyElement = document.getElementById("quantity-" + prod.id);
    if (qtyElement) {
      const currentQty = parseFloat(qtyElement.innerText.trim());
      if (currentQty) {
        qtyElement.innerText = currentQty - prod.quantity;
      }
    }
  }
}
