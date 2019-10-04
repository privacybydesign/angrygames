(function ($) {
	"use strict"

	// Mobile Nav toggle
	$('.menu-toggle > a').on('click', function (e) {
		e.preventDefault();
		$('#responsive-nav').toggleClass('active');
	})

	// Fix cart dropdown from closing
	$('.cart-dropdown').on('click', function (e) {
		e.stopPropagation();
	});

	/////////////////////////////////////////

	// Products Slick
	$('.products-slick').each(function () {
		var $this = $(this),
			$nav = $this.attr('data-nav');

		$this.slick({
			slidesToShow: 4,
			slidesToScroll: 1,
			autoplay: true,
			infinite: true,
			speed: 300,
			dots: false,
			arrows: true,
			appendArrows: $nav ? $nav : false,
			responsive: [{
					breakpoint: 991,
					settings: {
						slidesToShow: 2,
						slidesToScroll: 1,
					}
				},
				{
					breakpoint: 480,
					settings: {
						slidesToShow: 1,
						slidesToScroll: 1,
					}
				},
			]
		});
	});

	// Product Main img Slick
	$('#product-main-img').slick({
		infinite: true,
		speed: 300,
		dots: false,
		arrows: true,
		fade: true,
		asNavFor: '#product-imgs',
	});

	// Product imgs Slick
	$('#product-imgs').slick({
		slidesToShow: 3,
		slidesToScroll: 1,
		arrows: true,
		centerMode: true,
		focusOnSelect: true,
		centerPadding: 0,
		vertical: true,
		asNavFor: '#product-main-img',
		responsive: [{
			breakpoint: 991,
			settings: {
				vertical: false,
				arrows: false,
				dots: true,
			}
		}, ]
	});

	// Product img zoom
	var zoomMainProduct = document.getElementById('product-main-img');
	if (zoomMainProduct) {
		$('#product-main-img .product-preview').zoom();
	}

	// Age verification with IRMA
	$('.add-to-cart-btn').on('click', function (element) {
		$('#snackbar').removeClass('show');
		console.log("Age verification started");
		$.get('/start_session.php?type='+element.currentTarget.attr('data-minage')+'plus', function (sessionpackagejson) {
			let sessionpackage = JSON.parse(sessionpackagejson);
			let options = {
				server: sessionpackage.sessionPtr.u.split('/irma')[0],
				token: sessionpackage.token,
				language: 'nl'
			};
			let promise = irma.handleSession(sessionpackage.sessionPtr, options);

			let success = function (data) {
				console.log("Session successful!");
				console.log("Result:", data);
				// Continue to order page if user is 18+
				if (data.disclosed[0][0].rawvalue.toLowerCase() === 'yes') {
					window.location.href = 'cart.html?game=' + element.currentTarget.id;
				} else {
					$('#snackbar-content').html('U hebt niet de juiste leeftijd om dit artikel te mogen bestellen.');
					$('#snackbar').addClass('show');
				}
			};

			let error = function (data) {
				if (data === 'CANCELLED') {
					console.log("Session cancelled!");
					$('#snackbar-content').html('Leeftijdsverificatie via IRMA is geannuleerd.');
					$('#snackbar').addClass('show');
				} else {
					console.log("Session failed!");
					console.log("Error data:", data);
					$('#snackbar-content').html('Er is een fout opgetreden bij de leeftijdsverificatie via IRMA.');
					$('#snackbar').addClass('show');
				}
			};

			promise.then(success, error);
		});
	});

})(jQuery);
