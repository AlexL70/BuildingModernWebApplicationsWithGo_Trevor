const attention = Prompt();
(() => {
  'use strict'

  // Fetch all the forms we want to apply custom Bootstrap validation styles to
  const forms = document.querySelectorAll('.needs-validation')

  // Loop over them and prevent submission
  Array.from(forms).forEach(form => {
    form.addEventListener('submit', event => {
      if (!form.checkValidity()) {
        event.preventDefault()
        event.stopPropagation()
      }

      form.classList.add('was-validated')
    }, false)
  })
})()

function notify(message, msgType="success") {
  notie.alert({
    type: msgType, // optional, default = 4, enum: [1, 2, 3, 4, 5, 'success', 'warning', 'error', 'info', 'neutral']
    text: message,
    stay: false, // optional, default = false
    time: 2, // optional, default = 3, minimum = 1,
    position: "top" // optional, default = 'top', enum: ['top', 'bottom']
  });
}

function notifyModal(title, text, icon, confirmButtonText) {
  Swal.fire({
    title: title,
    html: text,
    icon: icon,
    confirmButtonText: confirmButtonText,
  })       
}

function Prompt() {
  let toast = function(c) {
    const {
      msg = "",
      icon = "success",
      position = "top-end"
    } = c;
    const Toast = Swal.mixin({
      toast: true,
      title: msg,
      position: position,
      icon: icon,
      showConfirmButton: false,
      timer: 3000,
      timerProgressBar: true,
      didOpen: (toast) => {
        toast.addEventListener('mouseenter', Swal.stopTimer)
        toast.addEventListener('mouseleave', Swal.resumeTimer)
      }
    })
    
    Toast.fire({})
  }

  let success = function(c) {
    const {
      msg = "",
      title = "",
      footer = "",

    } = c;
    Swal.fire({
      icon: 'success',
      title: title,
      text: msg,
      footer: footer,
    })
  }

  let error = function(c) {
    const {
      msg = "",
      title = "",
      footer = "",

    } = c;
    Swal.fire({
      icon: 'error',
      title: title,
      text: msg,
      footer: footer,
    })
  }

  let custom = async function(c) {
    const {
      icon = "",
      msg = "",
      title = "",
      showConfirmButton = true,
    } = c;

    const result = await Swal.fire({
      icon: icon,
      title: title,
      html: msg,
      backdrop: false,
      focusConfirm: false,
      showCancelButton: true,
      showConfirmButton: showConfirmButton,
      willOpen: () => {
        if (c.willOpen !== undefined) {
          c.willOpen();
        }
      },
      didOpen: () => {
        if (c.didOpen !== undefined) {
          c.didOpen();
        }
      },
    })

    if (result && result.isConfirmed && !result.isDismissed && result.value) {
      if (c.callback !== undefined) {
        c.callback(result.value);
      }
    }
  }

  let availability = function (roomID, csrf_token) {
      document.getElementById("check-availability-button").addEventListener("click", function(){
        let html = `
          <form id="check-availability-form" action="" method="post" novalidate class"needs-validation">
            <div class="row m-3" id="reservation-dates-modal">
              <div class="col">
                <label for="start" class="form-label">Arrival Date</label>
                <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival" autocomplete="off">
              </div>
              <div class="col">
                <label for="end" class="form-label">Departure Date</label>
                <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure" autocomplete="off">
              </div>
            </div>
          </form>
        `;

        custom({
          title: "Please, choose your dates:",
          msg: html,
          willOpen: () => {
            const elem = document.getElementById("check-availability-form");
            const rp = new DateRangePicker(elem, {
              format: "yyyy-mm-dd",
              autohide: true,
              showOnFocus: true,
              minDate: new Date(),
            })
          },
          didOpen: () => {
              document.getElementById("start").removeAttribute("disabled");
              document.getElementById("end").removeAttribute("disabled");
          },
          preConfirm: () => {
            return [
              document.getElementById("start").value,
              document.getElementById("end").value
            ]
          },
          callback: function(result) {
            let form = document.getElementById("check-availability-form");
            let formData = new FormData(form)
            formData.append("csrf_token", csrf_token)
            formData.append("room_id", roomID)

            fetch("/search-availability-json", {
              method: "POST",
              body: formData,
            })
              .then(response => response.json())
              .then(data => {
                if (data.ok) {
                  attention.custom({
                    icon: "success",
                    title: "Available!",
                    showConfirmButton: false,
                    msg: `
                    <p>Room is available for the time interval you've chosen.</p>
                    <p><a href="/book-room?id=${data.room_id}&start=${data.start_date}&end=${data.end_date}" class="btn btn-primary">Book now!</a></p>
                    `
                  });
                } else {
                  attention.error({title: "Sorry!", msg: "Room is not available for the time interval you've chosen."})
                }
              })
          }
        })
      });

  }

  return {
    toast: toast,
    success: success,
    error: error,
    custom: custom,
    availability: availability,
  }
}
